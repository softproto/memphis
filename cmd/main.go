package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"memphis/cmd/contractor"
	"memphis/cmd/inbox"
	"memphis/cmd/outbox"

	kafka "github.com/segmentio/kafka-go"
)

const inbox_topic = "inbox"
const inbox_partition = 0

const kafka_retries = 10
const kafka_addr = "kafka:29092"


func main() {
	log.Println("main() started")

	//
	var jobs = make(chan contractor.Job, contractor.WorkersCount)
	defer close(jobs)
	//
	ctx, contextCancel := context.WithCancel(context.Background())
	//
	inboxConn := connectToKafka(ctx, kafka_addr, inbox_topic, inbox_partition, kafka_retries)
	defer inboxConn.Close()
	//
	var wg sync.WaitGroup

	//продюссер в локальную кафку
	go producer(inboxConn)

	wg.Add(1)
	//Contractor (Подрядчик) - Обработчик сообщений.
	//Берет сообщения из "Входящих" (inbox) и отправляет результат в "Исходящие" (outbox)
	//Реализован пул воркеров (незавершаемые горутины - после выполнения задачи берут следующую из списка)
	go contractor.Run(ctx, &wg, jobs)

	//Входящие.
	//Читает сообщения из топика кафки и отправляет в Обработчик
	wg.Add(1)
	go inbox.Run(ctx, &wg, jobs)

	//Исходящие.
	//Получает результат от обработчика и записывает в топик кафки
	wg.Add(1)
	go outbox.Run(ctx, &wg, jobs)

	//Обрабатывает сигнал ОС на завершение
	gracefulShutDown()
	//Передаем сигнал о завершении всем потомкам
	contextCancel()

	wg.Wait()

	log.Println("main() exited")
}

func connectToKafka(ctx context.Context, addr string, topic string, partition int, retries int) *kafka.Conn {
	var conn *kafka.Conn
	var err error

	for i := 0; i < retries; i++ {
		log.Println("try ", i)
		conn, err = kafka.DialLeader(ctx, "tcp", addr, topic, partition)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

func producer(conn *kafka.Conn) {
	for {

		n := rand.Intn(100)

		err := conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
		if err != nil {
			log.Fatal(err)
		}

		// пишем сообщение в кафку
		_, err = conn.WriteMessages(
			kafka.Message{Value: []byte(strconv.Itoa(n))},
		)
		if err != nil {
			log.Fatal(err)
		}

		// log.Println("Written to topic:", n)

		time.Sleep(1 * time.Second)

	}

}

func gracefulShutDown() {

	ch := make(chan os.Signal, 2)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(ch)

	sig := <-ch
	log.Printf("%s %v - %s", "Received shutdown signal:", sig, "Graceful shutdown ... done")

}
