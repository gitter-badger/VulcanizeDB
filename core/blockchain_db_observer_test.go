package core_test

import (
	"math/big"

	"fmt"

	"github.com/8thlight/vulcanizedb/core"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "vulcanize"
)

var _ = Describe("Saving blocks to the database", func() {

	var db *sqlx.DB
	var err error
	pgConfig := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	BeforeEach(func() {
		db, err = sqlx.Connect("postgres", pgConfig)
		db.MustExec("DELETE FROM blocks")
		db.MustExec("DELETE FROM transactions")
	})

	It("implements the observer interface", func() {
		var observer core.BlockchainObserver = core.BlockchainDBObserver{Db: db}
		Expect(observer).NotTo(BeNil())
	})

	It("connects to the database", func() {
		Expect(err).Should(BeNil())
		Expect(db).ShouldNot(BeNil())
	})

	It("starts with no blocks", func() {
		var count int
		queryError := db.Get(&count, "SELECT COUNT(*) FROM blocks")
		Expect(queryError).Should(BeNil())
		Expect(count).Should(Equal(0))
	})

	It("inserts a block", func() {
		// setup a block in memory
		blockNumber := big.NewInt(1)
		gasLimit := big.NewInt(1000000)
		gasUsed := big.NewInt(10)
		blockTime := big.NewInt(1508981640)
		block := core.Block{Number: blockNumber, GasLimit: gasLimit, GasUsed: gasUsed, Time: blockTime}

		// save the block to the database
		observer := core.BlockchainDBObserver{Db: db}
		observer.NotifyBlockAdded(block)

		// find the saved block
		rows, err := db.Query("SELECT block_number, block_gaslimit, block_gasused, block_time FROM blocks")
		Expect(err).To(BeNil())
		var savedBlocks []core.Block
		for rows.Next() {
			var blockNumber int64
			var blockTime float64
			var gasLimit float64
			var gasUsed float64
			rows.Scan(&blockNumber, &gasLimit, &gasUsed, &blockTime)
			savedBlock := core.Block{
				GasUsed:  big.NewInt(int64(gasUsed)),
				GasLimit: big.NewInt(int64(gasLimit)),
				Number:   big.NewInt(blockNumber),
				Time:     big.NewInt(int64(blockTime)),
			}
			savedBlocks = append(savedBlocks, savedBlock)
		}
		// assert against the attributes
		Expect(len(savedBlocks)).To(Equal(1))
		Expect(savedBlocks[0].Number.Int64()).To(Equal(blockNumber.Int64()))
		Expect(savedBlocks[0].GasLimit.Int64()).To(Equal(gasLimit.Int64()))
		Expect(savedBlocks[0].GasUsed.Int64()).To(Equal(gasUsed.Int64()))
		Expect(savedBlocks[0].Time).To(Equal(blockTime))
	})

	var _ = Describe("Saving transactions to the database", func() {

		It("inserts a transaction", func() {
			gasLimit := int64(5000)
			gasPrice := int64(3)
			nonce := uint64(10000)
			to := "1234567890"
			value := int64(10)

			txRecord := core.Transaction{
				Hash:     "x1234",
				GasPrice: gasPrice,
				GasLimit: gasLimit,
				Nonce:    nonce,
				To:       to,
				Value:    value,
			}
			blockNumber := big.NewInt(1)
			gasUsed := big.NewInt(10)
			blockTime := big.NewInt(1508981640)
			block := core.Block{Number: blockNumber, GasLimit: big.NewInt(gasLimit), GasUsed: gasUsed, Time: blockTime, Transactions: []core.Transaction{txRecord}}

			observer := core.BlockchainDBObserver{Db: db}
			observer.NotifyBlockAdded(block)

			rows, err := db.Query("SELECT tx_hash, tx_nonce, tx_to, tx_gaslimit, tx_gasprice, tx_value FROM transactions")
			Expect(err).To(BeNil())

			var savedTransactions []core.Transaction
			for rows.Next() {
				var dbHash string
				var dbNonce uint64
				var dbTo string
				var dbGasLimit int64
				var dbGasPrice int64
				var dbValue int64
				rows.Scan(&dbHash, &dbNonce, &dbTo, &dbGasLimit, &dbGasPrice, &dbValue)
				savedTransaction := core.Transaction{
					Hash:     dbHash,
					Nonce:    dbNonce,
					To:       dbTo,
					GasLimit: dbGasLimit,
					GasPrice: dbGasPrice,
					Value:    dbValue,
				}
				savedTransactions = append(savedTransactions, savedTransaction)
			}

			Expect(len(savedTransactions)).To(Equal(1))
			savedTransaction := savedTransactions[0]
			Expect(savedTransaction.Hash).To(Equal(txRecord.Hash))
			Expect(savedTransaction.To).To(Equal(to))
			Expect(savedTransaction.Nonce).To(Equal(nonce))
			Expect(savedTransaction.GasLimit).To(Equal(gasLimit))
			Expect(savedTransaction.GasPrice).To(Equal(gasPrice))
			Expect(savedTransaction.Value).To(Equal(value))
		})
	})

})
