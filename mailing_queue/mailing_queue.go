package mailing_queue

import (
	"gitlab.com/michalSolarz/AuthAPI/authorization"
	"github.com/adjust/redismq"
	"encoding/json"
)

func QueueTransactionalMail(queue *redismq.Queue, token authorization.MailingToken) {
	t, err := json.Marshal(token)
	if err != nil {
		panic(err)
	}
	queue.Put(string(t))
}
