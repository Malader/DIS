package queue

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"CrackHash/manager/internal/config"
	"CrackHash/manager/internal/types"

	"github.com/rabbitmq/amqp091-go"
)

type TaskQueue interface {
	IsConnected() bool
	PublishTask(task types.CrackHashManagerRequest) error

	StartConsumeResponses() (<-chan types.CrackHashWorkerResponse, error)

	AckMessage(resp types.CrackHashWorkerResponse)
}

type rabbitClient struct {
	mu        sync.Mutex
	conn      *amqp091.Connection
	channel   *amqp091.Channel
	cfg       *config.Config
	connected bool

	exchangeTasks string
	exchangeResps string
	taskQueue     amqp091.Queue
	responseQueue amqp091.Queue

	responseChan <-chan amqp091.Delivery
	responsesOut chan types.CrackHashWorkerResponse
	listening    bool
}

func NewRabbitClient(cfg *config.Config) (TaskQueue, error) {
	client := &rabbitClient{
		cfg:           cfg,
		exchangeTasks: cfg.TaskExchange,
		exchangeResps: cfg.ResponseExchange,
		connected:     false,
		responsesOut:  make(chan types.CrackHashWorkerResponse),
	}

	if err := client.connectAndDeclare(); err != nil {
		return nil, err
	}
	client.startCloseWatcher()
	return client, nil
}

func (r *rabbitClient) connectAndDeclare() error {
	conn, err := amqp091.Dial(r.cfg.RabbitURI)
	if err != nil {
		return fmt.Errorf("не удалось подключиться к RabbitMQ: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("не удалось открыть канал: %w", err)
	}

	err = ch.ExchangeDeclare(
		r.cfg.TaskExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("не удалось объявить exchange задач: %w", err)
	}

	taskQ, err := ch.QueueDeclare(
		r.cfg.TaskQueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("не удалось объявить очередь задач: %w", err)
	}
	err = ch.QueueBind(
		taskQ.Name,
		taskQ.Name,
		r.cfg.TaskExchange,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("не удалось bind очередь задач: %w", err)
	}

	err = ch.ExchangeDeclare(
		r.cfg.ResponseExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("не удалось объявить exchange ответов: %w", err)
	}

	respQ, err := ch.QueueDeclare(
		r.cfg.ResponseQueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("не удалось объявить очередь ответов: %w", err)
	}
	err = ch.QueueBind(
		respQ.Name,
		respQ.Name,
		r.cfg.ResponseExchange,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("не удалось bind очередь ответов: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.conn = conn
	r.channel = ch
	r.taskQueue = taskQ
	r.responseQueue = respQ
	r.connected = true

	log.Println("[rabbitClient] Успешно подключился к RabbitMQ и объявил exchange/queues")
	return nil
}

func (r *rabbitClient) startCloseWatcher() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.conn == nil {
		return
	}
	closeCh := r.conn.NotifyClose(make(chan *amqp091.Error))

	go func() {
		err, ok := <-closeCh
		if !ok || err == nil {
			log.Println("[rabbitClient] NotifyClose: канал закрыт без ошибки")
			return
		}
		log.Printf("[rabbitClient] Соединение с RabbitMQ закрыто: %v", err)
		r.reconnectLoop()
	}()
}

func (r *rabbitClient) reconnectLoop() {
	r.mu.Lock()
	wasListening := r.listening
	r.connected = false
	r.mu.Unlock()

	for {
		log.Println("[rabbitClient] Пытаемся переподключиться к RabbitMQ...")
		err := r.connectAndDeclare()
		if err == nil {
			r.startCloseWatcher()
			log.Println("[rabbitClient] Успешно переподключились к RabbitMQ!")

			if wasListening {
				log.Println("[rabbitClient] Повторный вызов consumeResponses() после reconnect")
				_, err2 := r.consumeResponses()
				if err2 != nil {
					log.Printf("[rabbitClient] Ошибка при повторном consumeResponses: %v", err2)
				}
			}
			return
		}
		log.Printf("[rabbitClient] Ошибка при reconnect: %v. Ждем 5с и повторяем...", err)
		time.Sleep(5 * time.Second)
	}
}

func (r *rabbitClient) IsConnected() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.connected
}

func (r *rabbitClient) PublishTask(task types.CrackHashManagerRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.connected {
		return errors.New("RabbitMQ не подключен")
	}

	xmlData, err := xml.MarshalIndent(task, "", "  ")
	if err != nil {
		return err
	}
	xmlData = append([]byte(xml.Header), xmlData...)

	err = r.channel.Publish(
		r.exchangeTasks,
		r.taskQueue.Name,
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/xml",
			Body:         xmlData,
			DeliveryMode: amqp091.Persistent,
		},
	)
	if err != nil {
		return fmt.Errorf("ошибка при publish в task_queue: %w", err)
	}
	return nil
}

func (r *rabbitClient) StartConsumeResponses() (<-chan types.CrackHashWorkerResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.listening {
		return r.responsesOut, nil
	}
	if !r.connected {
		return nil, errors.New("RabbitMQ не подключен, попробуйте позже")
	}
	r.listening = true

	_, err := r.consumeResponses()
	if err != nil {
		return nil, err
	}
	return r.responsesOut, nil
}

func (r *rabbitClient) consumeResponses() (<-chan types.CrackHashWorkerResponse, error) {
	deliveries, err := r.channel.Consume(
		r.responseQueue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("не удалось подписаться на очередь %s: %w", r.responseQueue.Name, err)
	}
	r.responseChan = deliveries

	go func() {
		for msg := range deliveries {
			var workerResp types.CrackHashWorkerResponse
			if unErr := xml.Unmarshal(msg.Body, &workerResp); unErr != nil {
				log.Printf("Ошибка разбора XML из очереди worker_responses: %v", unErr)
				msg.Nack(false, false)
				continue
			}
			workerResp.DeliveryTag = msg.DeliveryTag
			r.responsesOut <- workerResp
		}
	}()

	return r.responsesOut, nil
}

func (r *rabbitClient) AckMessage(resp types.CrackHashWorkerResponse) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.connected {
		return
	}
	err := r.channel.Ack(resp.DeliveryTag, false)
	if err != nil {
		log.Printf("Ошибка при подтверждении сообщения RabbitMQ: %v", err)
	}
}
