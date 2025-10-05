package kafka_producer

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/IBM/sarama"

	"ep.k16/newsfeed/internal/service/model"
	"ep.k16/newsfeed/pkg/logger"
)

type KafkaProducer struct {
	cfg            KafkaConfig
	saramaProducer sarama.SyncProducer

	wg sync.WaitGroup
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
}

func New(cfg KafkaConfig) (*KafkaProducer, error) {

	// basic producer settings
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	saramaCfg.Producer.Partitioner = sarama.NewHashPartitioner // partition by key

	saramaProducer, err := sarama.NewSyncProducer(cfg.Brokers, saramaCfg)
	if err != nil {
		return nil, err
	}

	p := &KafkaProducer{
		cfg:            cfg,
		saramaProducer: saramaProducer,
	}

	logger.Info("init kafka producer successfully", logger.F("cfg", cfg))
	return p, nil
}

func (p *KafkaProducer) Stop() {
	// wait until all ongoing messages are done, then we shut down the producer
	p.wg.Wait()

	p.saramaProducer.Close()
}

func (p *KafkaProducer) SendPost(ctx context.Context, post *model.Post) error {
	p.wg.Add(1)
	defer p.wg.Done()

	data, err := json.Marshal(post)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: p.cfg.Topic,
		// key is anything u want, here use user_id, so all posts from the same grpc will be consumed by the same instance
		Key:       sarama.StringEncoder(strconv.Itoa(int(post.UserID))),
		Value:     sarama.ByteEncoder(data),
		Timestamp: time.Now(),
	}

	partition, offset, err := p.saramaProducer.SendMessage(msg)
	if err != nil {
		return err
	}
	logger.Debug("send msg successfully",
		logger.F("partition", partition),
		logger.F("offset", offset),
		logger.F("message", string(data)))

	return nil
}
