package newsfeed_processor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/IBM/sarama"

	"ep.k16/newsfeed/internal/service/model"
	"ep.k16/newsfeed/pkg/logger"
)

type NewsfeedService interface {
	AppendPostToNewsfeed(ctx context.Context, post *model.Post) error
}

type NewsfeedBuilder struct {
	cfg Config

	saramaConsumer sarama.ConsumerGroup

	handler *postMsgHandler
}

type Config struct {
	Brokers       []string
	Topic         string
	ConsumerGroup string
}

func New(cfg Config, newsfeedService NewsfeedService) (*NewsfeedBuilder, error) {

	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	// create kafka consumer
	consumerGroup, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.ConsumerGroup, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %s", err)
	}

	// create post message handler
	handler := &postMsgHandler{
		newsfeedService: newsfeedService,
		ready:           make(chan bool),
	}

	return &NewsfeedBuilder{
		cfg:            cfg,
		saramaConsumer: consumerGroup,
		handler:        handler,
	}, nil
}

func (p *NewsfeedBuilder) Start(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := p.saramaConsumer.Consume(ctx, []string{p.cfg.Topic}, p.handler); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				logger.Error("Error from consumer", logger.E(err))
				errCh <- err
				return
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				errCh <- ctx.Err()
				return
			}
			p.handler.ready = make(chan bool) // TODO: ????
		}
	}()

	select {
	case <-p.handler.ready: // Await till the consumer has been set up
	case err := <-errCh:
		return err
	}

	logger.Info("sarama consumer up and running!...")

	<-ctx.Done()
	logger.Info("sarama stopped due to context canceled")
	return nil
}

func (p *NewsfeedBuilder) Stop() {
	p.saramaConsumer.Close()
}

// follow demo here: https://github.com/IBM/sarama/blob/main/examples/consumergroup/main.go
// postMsgHandler represents a Sarama consumer group consumer  (follows sarama.ConsumerGroupHandler interface)
type postMsgHandler struct {
	newsfeedService NewsfeedService

	ready chan bool
}

// Setup is called at the beginning of a new session, before ConsumeClaim
func (h *postMsgHandler) Setup(sarama.ConsumerGroupSession) error {
	logger.Info("postMsgHandler setting up")
	close(h.ready)
	return nil
}

// Cleanup is called at the end of a session, once all ConsumeClaim goroutines have exited
func (h *postMsgHandler) Cleanup(sarama.ConsumerGroupSession) error {
	logger.Info("postMsgHandler cleaning up")
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing loop and exit.
func (h *postMsgHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	logger.Info("postMsgHandler consuming ...")
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				logger.Info("postMsgHandler message channel was closed")
				return nil
			}

			start := time.Now()
			logFields := []logger.Field{
				logger.F("value", string(message.Value)),
				logger.F("topic", message.Topic),
				logger.F("ts", message.Timestamp),
			}

			err := h.newsfeedService.AppendPostToNewsfeed(context.Background(), &model.Post{})
			if err != nil {
				logFields = append(logFields, logger.E(err), logger.F("latency", time.Since(start)))
				logger.Error("failed to process message", logFields...)
			} else {
				logFields = append(logFields, logger.F("latency", time.Since(start)))
				logger.Info("processed message", logFields...)
			}

			session.MarkMessage(message, "")
		case <-session.Context().Done():
			logger.Info("postMsgHandler stopped due to context canceled")
			return nil
		}
	}
}
