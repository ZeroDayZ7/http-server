package worker

import "time"

const (
	streamName     = "interaction_events"
	dlqStream      = "interaction_events_dlq"
	groupName      = "interaction_workers"
	idemKey        = "interaction_events_idem"
	idemTTL        = 24 * time.Hour
	readBlockTime  = 2 * time.Second
	readBatchSize  = 500
	batchSizeLimit = 2000

	reclaimInterval   = 1000 * time.Millisecond
	minIdleTime       = 500 * time.Millisecond
	maxReclaimPerLoop = 500

	flushTimeout   = 5 * time.Second
	ackTimeout     = 3 * time.Second
	moveDLQTimeout = 3 * time.Second

	maxFlushRetries = 3
	maxAckRetries   = 3

	msgChanSize = batchSizeLimit * 4
)
