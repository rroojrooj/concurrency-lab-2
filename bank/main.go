package main

import (
	"container/list"
	"flag"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

var debug *bool

// readyQueue will be used by the manager to push transactions ready for execution
var readyQueue chan transaction

func manager(bank *bank, transactionQueue <-chan transaction, readyQueue chan<- transaction) {
	for t := range transactionQueue {
		for {
			fromLocked := bank.isAccountLocked(t.from)
			toLocked := bank.isAccountLocked(t.to)
			if !fromLocked && !toLocked {
				bank.lockAccount(t.from, "Manager")
				if bank.isAccountLocked(t.to) {
					bank.unlockAccount(t.from, "Manager")
					time.Sleep(10 * time.Millisecond) // Small sleep to avoid busy-waiting.
				} else {
					bank.lockAccount(t.to, "Manager")
					readyQueue <- t
					break
				}
			} else {
				time.Sleep(10 * time.Millisecond) // Small sleep to avoid busy-waiting.
			}
		}
	}
	close(readyQueue) // Close the readyQueue when all transactions have been processed
}

// An executor is a type of a worker goroutine that handles the incoming transactions.
func executor(bank *bank, executorId int, transactionQueue <-chan transaction, done chan<- bool) {
	for t := range readyQueue {
		from := bank.getAccountName(t.from)
		to := bank.getAccountName(t.to)

		fmt.Println("Executor\t", executorId, "attempting transaction from", from, "to", to)
		e := bank.addInProgress(t, executorId) // Removing this line will break visualisations.

		bank.execute(t, executorId)

		bank.unlockAccount(t.from, "Executor "+strconv.Itoa(executorId))
		bank.unlockAccount(t.to, "Executor "+strconv.Itoa(executorId))

		bank.removeCompleted(e, executorId) // Removing this line will break visualisations.
		done <- true
	}
}

func toChar(i int) rune {
	return rune('A' + i)
}

// main creates a bank and executors that will be handling the incoming transactions.
func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	debug = flag.Bool("debug", false, "generate DOT graphs of the state of the bank")
	flag.Parse()

	bankSize := 6 // Must be even for correct visualisation.
	transactions := 1000

	accounts := make([]*account, bankSize)
	for i := range accounts {
		accounts[i] = &account{name: string(toChar(i)), balance: 1000}
	}

	bank := bank{
		accounts:               accounts,
		transactionsInProgress: list.New(),
		gen:                    newGenerator(),
	}

	startSum := bank.sum()

	transactionQueue := make(chan transaction, transactions)
	readyQueue := make(chan transaction, transactions) // New ready queue

	expectedMoneyTransferred := 0
	for i := 0; i < transactions; i++ {
		t := bank.getTransaction()
		expectedMoneyTransferred += t.amount
		transactionQueue <- t
	}

	close(transactionQueue) // Close the transactionQueue after feeding all transactions

	// Start the manager
	go manager(&bank, transactionQueue, readyQueue) // Pass readyQueue to the manager as well

	done := make(chan bool)

	// Executors now listen to readyQueue
	for i := 0; i < bankSize; i++ {
		go executor(&bank, i, readyQueue, done)
	}

	for total := 0; total < transactions; total++ {
		fmt.Println("Completed transactions\t", total)
		<-done
	}

	fmt.Println()
	fmt.Println("Expected transferred", expectedMoneyTransferred)
	fmt.Println("Actual transferred", bank.moneyTransferred)
	fmt.Println("Expected sum", startSum)
	fmt.Println("Actual sum", bank.sum())
	if bank.sum() != startSum {
		panic("sum of the account balances does not match the starting sum")
	} else if bank.moneyTransferred != expectedMoneyTransferred {
		panic("incorrect amount of money was transferred")
	} else {
		fmt.Println("The bank works!")
	}
}
