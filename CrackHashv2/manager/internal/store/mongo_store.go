package store

import (
	"context"
	"log"
	"time"

	"CrackHash/manager/internal/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRequestStore struct {
	client     *mongo.Client
	collection *mongo.Collection
}

type RequestDocument struct {
	ID        string   `bson:"_id"`
	Status    string   `bson:"status"`
	Data      []string `bson:"data"`
	StartTime time.Time
	Timeout   time.Duration
	Pending   bool
	Hash      string
	MaxLength int
}

func NewMongoRequestStore(cfg *config.Config) (*MongoRequestStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return nil, err
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	store := &MongoRequestStore{
		client:     client,
		collection: client.Database(cfg.MongoDatabase).Collection("requests"),
	}
	return store, nil
}

func (m *MongoRequestStore) Set(id string, state RequestState) {
	doc := RequestDocument{
		ID:        id,
		Status:    state.Status,
		Data:      state.Data,
		StartTime: state.StartTime,
		Timeout:   state.Timeout,
		Pending:   false,
		Hash:      "",
		MaxLength: 0,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	opts := options.Update().SetUpsert(true)
	_, err := m.collection.UpdateByID(ctx, id, bson.M{"$set": doc}, opts)
	if err != nil {
		log.Printf("Ошибка при сохранении Set: %v", err)
	}
}

func (m *MongoRequestStore) Get(id string) (RequestState, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var doc RequestDocument
	err := m.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		return RequestState{}, false
	}
	state := RequestState{
		Status:    doc.Status,
		Data:      doc.Data,
		StartTime: doc.StartTime,
		Timeout:   doc.Timeout,
		Timer:     nil,
	}
	return state, true
}

func (m *MongoRequestStore) Update(id string, state RequestState) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	update := bson.M{
		"status":    state.Status,
		"data":      state.Data,
		"starttime": state.StartTime,
		"timeout":   state.Timeout,
	}
	_, err := m.collection.UpdateByID(ctx, id, bson.M{"$set": update})
	if err != nil {
		log.Printf("Ошибка при Update: %v", err)
	}
}

func (m *MongoRequestStore) Count() int {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	count, err := m.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Printf("Ошибка при Count: %v", err)
		return 0
	}
	return int(count)
}

func (m *MongoRequestStore) MarkPending(id string, isPending bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := m.collection.UpdateByID(ctx, id, bson.M{"$set": bson.M{"pending": isPending}})
	if err != nil {
		log.Printf("Ошибка при MarkPending(%s): %v", id, err)
	}
}

func (m *MongoRequestStore) GetPending() []PendingTask {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := m.collection.Find(ctx, bson.M{"pending": true})
	if err != nil {
		log.Printf("Ошибка при GetPending: %v", err)
		return nil
	}
	defer cursor.Close(ctx)

	var docs []RequestDocument
	if err = cursor.All(ctx, &docs); err != nil {
		log.Printf("Ошибка при cursor.All в GetPending: %v", err)
		return nil
	}

	var result []PendingTask
	for _, d := range docs {
		result = append(result, PendingTask{
			ID:        d.ID,
			Hash:      d.Hash,
			MaxLength: d.MaxLength,
			Status:    d.Status,
		})
	}
	return result
}
