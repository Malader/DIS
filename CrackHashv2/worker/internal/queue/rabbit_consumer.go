package queue

import (
	"CrackHash/worker/internal/config"
	"CrackHash/worker/internal/handlers"
	"CrackHash/worker/internal/types"
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type TaskConsumer interface {
	StartConsuming(ctx context.Context, svc handlers.WorkerService) error
}

type rabbitConsumer struct {
	mu        sync.Mutex
	conn      *amqp091.Connection
	channel   *amqp091.Channel
	cfg       config.Config
	connected bool

	taskQueueName string
	respQueueName string

	svc handlers.WorkerService
	ctx context.Context

	consuming bool
}

func NewRabbitConsumer(cfg config.Config) (TaskConsumer, error) {
	c := &rabbitConsumer{
		cfg:       cfg,
		connected: false,
	}
	if err := c.connectAndDeclare(); err != nil {
		return nil, err
	}
	c.startCloseWatcher()
	return c, nil
}

func (r *rabbitConsumer) StartConsuming(ctx context.Context, svc handlers.WorkerService) error {
	r.mu.Lock()
	r.ctx = ctx
	r.svc = svc
	if r.consuming {
		r.mu.Unlock()
		return nil
	}
	r.consuming = true
	r.mu.Unlock()

	return r.consumeTasks()
}

func (r *rabbitConsumer) consumeTasks() error {
	r.mu.Lock()
	if !r.connected {
		r.mu.Unlock()
		return fmt.Errorf("rabbitConsumer не подключен")
	}
	ch := r.channel
	queueName := r.taskQueueName
	ctx := r.ctx
	svc := r.svc
	r.mu.Unlock()

	msgs, err := ch.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("не удалось подписаться на очередь %s: %w", queueName, err)
	}

	go func() {
		log.Printf("[rabbitConsumer] Старт consumeTasks для %s", queueName)
		for {
			select {
			case <-ctx.Done():
				log.Println("[rabbitConsumer] Context done, останавливаем consumeTasks...")
				return
			case d, ok := <-msgs:
				if !ok {
					log.Println("[rabbitConsumer] Канал msgs закрыт, выходим из consumeTasks...")
					return
				}
				var req types.CrackHashManagerRequest
				if err := xml.Unmarshal(d.Body, &req); err != nil {
					log.Printf("[rabbitConsumer] Ошибка парсинга XML: %v", err)
					d.Nack(false, false)
					continue
				}
				log.Printf("[rabbitConsumer] Получили задачу: requestID=%s, hash=%s, maxLength=%d, partNumber=%d, partCount=%d",
					req.RequestId, req.Hash, req.MaxLength, req.PartNumber, req.PartCount)

				results := svc.ProcessTask(req.Hash, req.MaxLength, req.Alphabet.Symbols, req.PartNumber, req.PartCount)

				workerResp := types.CrackHashWorkerResponse{
					RequestId:  req.RequestId,
					PartNumber: req.PartNumber,
				}
				workerResp.Answers.Words = results

				if pubErr := r.publishResponse(workerResp); pubErr != nil {
					log.Printf("[rabbitConsumer] Ошибка при отправке ответа requestID=%s: %v", req.RequestId, pubErr)
					d.Nack(false, true)
					continue
				}
				log.Printf("[rabbitConsumer] Отправили ответ по requestID=%s, найдено слов=%d", req.RequestId, len(results))
				d.Ack(false)
			}
		}
	}()
	return nil
}

func (r *rabbitConsumer) connectAndDeclare() error {
	conn, err := amqp091.Dial(r.cfg.RabbitURI)
	if err != nil {
		return fmt.Errorf("не удалось подключиться к RabbitMQ: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("не удалось открыть канал: %w", err)
	}

	if err := ch.ExchangeDeclare(
		r.cfg.TaskExchange, "direct",
		true, false, false, false, nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("не удалось объявить exchange %s: %w", r.cfg.TaskExchange, err)
	}
	taskQ, err := ch.QueueDeclare(
		r.cfg.TaskQueueName,
		true, false, false, false, nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("не удалось объявить очередь %s: %w", r.cfg.TaskQueueName, err)
	}
	if err := ch.QueueBind(
		taskQ.Name, taskQ.Name,
		r.cfg.TaskExchange, false, nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("не удалось bind %s: %w", taskQ.Name, err)
	}

	if err := ch.ExchangeDeclare(
		r.cfg.ResponseExchange, "direct",
		true, false, false, false, nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("не удалось объявить exchange %s: %w", r.cfg.ResponseExchange, err)
	}
	respQ, err := ch.QueueDeclare(
		r.cfg.ResponseQueue,
		true, false, false, false, nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("не удалось объявить очередь %s: %w", r.cfg.ResponseQueue, err)
	}
	if err := ch.QueueBind(
		respQ.Name, respQ.Name,
		r.cfg.ResponseExchange, false, nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("не удалось bind %s: %w", respQ.Name, err)
	}

	r.mu.Lock()
	r.conn = conn
	r.channel = ch
	r.taskQueueName = taskQ.Name
	r.respQueueName = respQ.Name
	r.connected = true
	r.mu.Unlock()

	log.Println("[rabbitConsumer] Успешно подключились к RabbitMQ и объявили exchange/queue")
	return nil
}

func (r *rabbitConsumer) startCloseWatcher() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.conn == nil {
		return
	}
	closeCh := r.conn.NotifyClose(make(chan *amqp091.Error))
	go func() {
		err, ok := <-closeCh
		if !ok || err == nil {
			log.Println("[rabbitConsumer] NotifyClose: канал закрыт без ошибки")
			return
		}
		log.Printf("[rabbitConsumer] Соединение закрыто: %v", err)
		r.reconnectLoop()
	}()
}

func (r *rabbitConsumer) reconnectLoop() {
	r.mu.Lock()
	r.connected = false
	alreadyConsuming := r.consuming
	r.mu.Unlock()

	for {
		log.Println("[rabbitConsumer] Пытаемся переподключиться к RabbitMQ...")
		if err := r.connectAndDeclare(); err == nil {
			r.startCloseWatcher()
			log.Println("[rabbitConsumer] Успешно переподключились к RabbitMQ!")
			if alreadyConsuming {
				log.Println("[rabbitConsumer] Повторный вызов consumeTasks() после reconnect")
				if err2 := r.consumeTasks(); err2 != nil {
					log.Printf("[rabbitConsumer] Ошибка при consumeTasks: %v", err2)
				}
			}
			return
		}
		log.Printf("[rabbitConsumer] Ошибка при reconnect. Ждём 5с...")
		time.Sleep(5 * time.Second)
	}
}

func (r *rabbitConsumer) publishResponse(resp types.CrackHashWorkerResponse) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.connected {
		return fmt.Errorf("rabbitConsumer не подключен")
	}

	xmlData, err := xml.MarshalIndent(resp, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка маршалинга ответа: %w", err)
	}
	xmlData = append([]byte(xml.Header), xmlData...)

	log.Printf("[rabbitConsumer] Публикуем ответ в queue=%s (requestID=%s)", r.cfg.ResponseQueue, resp.RequestId)
	if err := r.channel.Publish(
		r.cfg.ResponseExchange,
		r.cfg.ResponseQueue,
		false, false,
		amqp091.Publishing{
			ContentType:  "application/xml",
			Body:         xmlData,
			DeliveryMode: amqp091.Persistent,
		},
	); err != nil {
		return fmt.Errorf("publishResponse error: %w", err)
	}
	return nil
}
