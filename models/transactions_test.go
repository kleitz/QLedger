package models

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TransactionsModelSuite struct {
	suite.Suite
	db *sql.DB
}

func (ts *TransactionsModelSuite) SetupSuite() {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	assert.NotEmpty(ts.T(), databaseURL)
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Panic("Unable to connect to Database:", err)
	} else {
		log.Println("Successfully established connection to database.")
		ts.db = db
	}
}

func (ts *TransactionsModelSuite) TestIsValid() {
	t := ts.T()

	transaction := &Transaction{
		ID: "t001",
		Lines: []*TransactionLine{
			&TransactionLine{
				AccountID: "a1",
				Delta:     100,
			},
			&TransactionLine{
				AccountID: "a2",
				Delta:     -100,
			},
		},
	}
	valid := transaction.IsValid()
	assert.Equal(t, valid, true, "Transaction should be valid")

	transaction.Lines[0].Delta = 200
	valid = transaction.IsValid()
	assert.Equal(t, valid, false, "Transaction should not be valid")
}

func (ts *TransactionsModelSuite) TestIsExists() {
	t := ts.T()

	transactionDB := TransactionDB{DB: ts.db}
	exists := transactionDB.IsExists("t001")
	assert.Equal(t, exists, false, "Transaction should not exist")

	transaction := &Transaction{
		ID: "t001",
		Lines: []*TransactionLine{
			&TransactionLine{
				AccountID: "a1",
				Delta:     100,
			},
			&TransactionLine{
				AccountID: "a2",
				Delta:     -100,
			},
		},
	}
	done := transactionDB.Transact(transaction)
	assert.Equal(t, done, true, "Transaction should be created")

	exists = transactionDB.IsExists("t001")
	assert.Equal(t, exists, true, "Transaction should exist")
}

func (ts *TransactionsModelSuite) TestIsConflict() {
	t := ts.T()

	transactionDB := TransactionDB{DB: ts.db}
	transaction := &Transaction{
		ID: "t002",
		Lines: []*TransactionLine{
			&TransactionLine{
				AccountID: "a1",
				Delta:     100,
			},
			&TransactionLine{
				AccountID: "a2",
				Delta:     -100,
			},
		},
	}
	done := transactionDB.Transact(transaction)
	assert.Equal(t, done, true, "Transaction should be created")

	conflicts := transactionDB.IsConflict(transaction)
	assert.Equal(t, conflicts, false, "Transaction should not conflict")

	transaction = &Transaction{
		ID: "t002",
		Lines: []*TransactionLine{
			&TransactionLine{
				AccountID: "a1",
				Delta:     50,
			},
			&TransactionLine{
				AccountID: "a2",
				Delta:     -50,
			},
		},
	}
	conflicts = transactionDB.IsConflict(transaction)
	assert.Equal(t, conflicts, true, "Transaction should conflict since deltas are different from first received")

	transaction = &Transaction{
		ID: "t002",
		Lines: []*TransactionLine{
			&TransactionLine{
				AccountID: "b1",
				Delta:     100,
			},
			&TransactionLine{
				AccountID: "b2",
				Delta:     -100,
			},
		},
	}
	conflicts = transactionDB.IsConflict(transaction)
	assert.Equal(t, conflicts, true, "Transaction should conflict since accounts are different from first received")
}

func (ts *TransactionsModelSuite) TestTransact() {
	t := ts.T()

	transactionDB := TransactionDB{DB: ts.db}

	transaction := &Transaction{
		ID: "t003",
		Lines: []*TransactionLine{
			&TransactionLine{
				AccountID: "a1",
				Delta:     100,
			},
			&TransactionLine{
				AccountID: "a2",
				Delta:     -100,
			},
		},
		Data: map[string]interface{}{
			"tag1": "val1",
			"tag2": "val2",
		},
	}
	done := transactionDB.Transact(transaction)
	assert.Equal(t, done, true, "Transaction should be created")

	exists := transactionDB.IsExists("t003")
	assert.Equal(t, exists, true, "Transaction should exist")

	done = transactionDB.Transact(transaction)
	assert.Equal(t, done, false, "Transaction should not be created")
}

func (ts *TransactionsModelSuite) TearDownSuite() {
	log.Println("Cleaning up the test database")

	t := ts.T()
	_, err := ts.db.Exec(`DELETE FROM lines`)
	if err != nil {
		t.Fatal("Error deleting lines:", err)
	}
	_, err = ts.db.Exec(`DELETE FROM transactions`)
	if err != nil {
		t.Fatal("Error deleting transactions:", err)
	}
	_, err = ts.db.Exec(`DELETE FROM accounts`)
	if err != nil {
		t.Fatal("Error deleting accounts:", err)
	}
}

func TestTransactionsModelSuite(t *testing.T) {
	suite.Run(t, new(TransactionsModelSuite))
}
