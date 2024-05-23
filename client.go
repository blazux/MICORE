package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

type Task struct {
	ID       string
	Command  string
	Interval time.Duration
	Stop     chan bool
	Repeat   int
}

type Message struct {
	Type   string // "task", "stop", or "list"
	Task   *Task
	TaskID string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: task_cli <command> [arguments...]")
		return
	}

	command := os.Args[1]

	switch command {
	case "add":
		if len(os.Args) != 5 {
			fmt.Println("Usage: task_cli add <command> <interval> <repeat>")
			return
		}
		addTask(os.Args[2], os.Args[3], os.Args[4])
	case "stop":
		if len(os.Args) != 3 {
			fmt.Println("Usage: task_cli stop <task_id>")
			return
		}
		stopTask(os.Args[2])
	case "list":
		listTasks()
	default:
		fmt.Println("Unknown command:", command)
	}
}

func addTask(command, intervalStr, repeatStr string) {
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		fmt.Println("Invalid interval:", err)
		return
	}

	repeat, err := strconv.Atoi(repeatStr)
	if err != nil {
		fmt.Println("Invalid repeat count:", err)
		return
	}

	task := &Task{
		ID:       generateID(),
		Command:  command,
		Interval: interval,
		Stop:     make(chan bool),
		Repeat:   repeat,
	}

	sendMessageToService(&Message{Type: "task", Task: task})
}

func stopTask(taskId string) {
	sendMessageToService(&Message{Type: "stop", TaskID: taskId})
}

func listTasks() {
	conn, err := net.Dial("tcp", "localhost:54030")
	if err != nil {
		fmt.Println("Failed to connect to service:", err)
		return
	}
	defer conn.Close()

	encoder := gob.NewEncoder(conn)
	err = encoder.Encode(&Message{Type: "list"})
	if err != nil {
		fmt.Println("Failed to send list request:", err)
		return
	}

	decoder := gob.NewDecoder(conn)
	var taskList []string
	err = decoder.Decode(&taskList)
	if err != nil {
		fmt.Println("Failed to receive task list:", err)
		return
	}

	fmt.Println("Running tasks:")
	for _, taskId := range taskList {
		fmt.Println(taskId)
	}
}

func sendMessageToService(msg *Message) {
	conn, err := net.Dial("tcp", "localhost:54030")
	if err != nil {
		fmt.Println("Failed to connect to service:", err)
		return
	}
	defer conn.Close()

	encoder := gob.NewEncoder(conn)
	fmt.Println("Sending message to service:", msg)
	err = encoder.Encode(msg)
	if err != nil {
		fmt.Println("Failed to send message to service:", err)
	} else {
		if msg.Type == "task" {
			fmt.Printf("Task %s added\n", msg.Task.ID)
		} else if msg.Type == "stop" {
			fmt.Printf("Task %s stop request sent\n", msg.TaskID)
		}
	}
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
