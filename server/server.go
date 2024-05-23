package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"os/exec"
	"sync"
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

var tasks = make(map[string]*Task)
var mu sync.Mutex

func main() {
	taskChan := make(chan *Task)
	stopChan := make(chan string)

	go taskManager(taskChan, stopChan)

	fmt.Println("Service started, waiting for tasks...")

	listener, err := net.Listen("tcp", "localhost:54030")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn, taskChan, stopChan)
	}
}

func handleConnection(conn net.Conn, taskChan chan *Task, stopChan chan string) {
	defer conn.Close()
	decoder := gob.NewDecoder(conn)
	var msg Message

	err := decoder.Decode(&msg)
	if err != nil {
		fmt.Println("Failed to decode message:", err)
		return
	}

	switch msg.Type {
	case "task":
		if msg.Task != nil {
			msg.Task.Stop = make(chan bool)
			fmt.Println("Decoded task:", *msg.Task)
			mu.Lock()
			tasks[msg.Task.ID] = msg.Task
			mu.Unlock()
			taskChan <- msg.Task
		}
	case "stop":
		fmt.Println("Decoded task ID:", msg.TaskID)
		stopChan <- msg.TaskID
	case "list":
		listTasks(conn)
	default:
		fmt.Println("Unknown message type:", msg.Type)
	}
}

func listTasks(conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()
	var taskList []string
	for id := range tasks {
		taskList = append(taskList, id)
	}
	encoder := gob.NewEncoder(conn)
	err := encoder.Encode(taskList)
	if err != nil {
		fmt.Println("Failed to send task list:", err)
	}
}

func taskManager(taskChan chan *Task, stopChan chan string) {
	for {
		select {
		case task := <-taskChan:
			go runTask(task)
		case taskId := <-stopChan:
			stopTask(taskId)
		}
	}
}

func runTask(task *Task) {
	ticker := time.NewTicker(task.Interval)
	defer ticker.Stop()
	fmt.Printf("Running task %s with command %s\n", task.ID, task.Command)
	for {
		select {
		case <-ticker.C:
			if task.Repeat == 0 {
				stopTask(task.ID)
				return
			}
			if task.Repeat > 0 {
				task.Repeat--
			}
			executeCommand(task.Command)
		case <-task.Stop:
			fmt.Printf("Stopping task %s\n", task.ID)
			return
		}
	}
}

func executeCommand(command string) {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error executing command: %v\n", err)
	} else {
		fmt.Printf("Output: %s\n", output)
	}
}

func stopTask(taskId string) {
	mu.Lock()
	defer mu.Unlock()
	if task, exists := tasks[taskId]; exists {
		fmt.Printf("Stopping task %s\n", taskId)
		close(task.Stop)
		delete(tasks, taskId)
	} else {
		fmt.Printf("Task %s not found\n", taskId)
	}
}
