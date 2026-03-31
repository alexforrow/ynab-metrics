package accounts

import (
	"fmt"
	"log"
	"strconv"
	"time"

	u "github.com/hoenn/ynab-metrics/pkg/units"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/brunomvsouza/ynab.go"
	"github.com/brunomvsouza/ynab.go/api/budget"
	"github.com/brunomvsouza/ynab.go/api/transaction"
	"github.com/brunomvsouza/ynab.go/api"
)

var accountBalance = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "account_balance",
	Help: "Account balance gauge",
},
	[]string{"budget_name", "name", "type", "closed"})

var attentionTransactions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "attention_transactions",
	Help: "Transactions requiring attention gauge",
},
	[]string{"budget_name", "account_name", "status"})

func init() {
	prometheus.MustRegister(accountBalance)
	prometheus.MustRegister(attentionTransactions)
}

//StartMetrics collects accounts metrics given a list of budgets
func StartMetrics(c ynab.ClientServicer, budgets []*budget.Budget) {
	log.Println("Getting Accounts...")

	statuses := map[string]*transaction.Status{
		"uncategorized": transaction.StatusUncategorized.Pointer(),
		"unapproved": transaction.StatusUnapproved.Pointer(),
	}

	// limit transactions to X months ago
	startDate := time.Now().AddDate(0, -1 ,0)
	startDateStr := startDate.Format(time.DateOnly)
	log.Print(fmt.Sprintf("Start date is: %s", startDateStr))

	for _, b := range budgets {
		for _, a := range b.Accounts {
			// output balance
			accountBalance.WithLabelValues(b.Name, a.Name, string(a.Type), strconv.FormatBool(a.Closed)).Set(float64(u.Dollars(a.Balance)))

			// dont do anything else for unused accounts
			if a.Closed || a.Deleted {
				log.Print(fmt.Sprintf("Skipping account %s/%s as its closed or deleted", b.Name, a.Name))
				continue
			}

			// for each account count number of transactions needing attention (unapproved or uncategorised)
			startDate, _ := api.DateFromString(startDateStr)

			for status, statusPointer := range statuses {
				filter := &transaction.Filter{
					Since: &startDate,
					Type:  statusPointer,
				}

				transactions, err := c.Transaction().GetTransactionsByAccount(
					b.ID, a.ID, filter)

				if err != nil {
					log.Print(fmt.Sprintf("Unable to get %s transactions for budget/account: %s/%s (%s)", status, b.Name, a.Name, err))
					continue
				}

				attentionTransactions.WithLabelValues(b.Name, a.Name, status).Set(float64(len(transactions)))
			}
		}
	}
}
