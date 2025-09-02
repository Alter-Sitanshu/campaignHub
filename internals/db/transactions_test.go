package db

import (
	"context"
	"log"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

var MockTsStore TransactionStore

func init() {
	MockTsStore.db = MockDB
}

func generateAccounts(ctx context.Context, holder_id, type_ string) *Account {
	acc_id := uuid.New().String()
	acc := Account{
		Id:       acc_id,
		HolderId: holder_id,
		Type:     type_,
		Amount:   10000.0,
	}
	err := MockTsStore.OpenAccount(ctx, &acc)
	if err != nil {
		log.Printf("error opening account: %v", err.Error())
		return nil
	}
	return &acc
}

func destroyAccounts(ctx context.Context, args ...string) {
	log.Printf("destroying account\n")
	for _, v := range args {
		MockTsStore.DeleteAccount(ctx, v)
	}
}

func destroyAllTransactions() {
	query := `DELETE FROM transactions`
	MockTsStore.db.Exec(query)
}

func TestPayout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	const amount float64 = 1000.0
	// create mock creator
	uid := "user1"
	generateCreator(ctx, uid)
	defer destroyCreator(ctx, uid)

	// create a mock brand
	bid := uuid.New().String()
	generateBrand(bid)
	defer destroyBrand(bid)

	// Create their accounts
	user_acc := generateAccounts(ctx, uid, "user")
	brand_acc := generateAccounts(ctx, bid, "brand")
	defer destroyAccounts(ctx, user_acc.Id, brand_acc.Id)

	// track the updated balances
	user_last_amount := user_acc.Amount
	brand_last_amount := brand_acc.Amount
	t.Run("making a single payment", func(t *testing.T) {
		invoice := Transaction{
			Id:     uuid.New().String(),
			FromId: brand_acc.Id,
			ToId:   user_acc.Id,
			Amount: amount,
			Type:   "payout",
		}
		err := MockTsStore.Payout(ctx, &invoice)
		if err != nil {
			log.Printf("error: %s\n", err.Error())
			t.Fail()
			return
		}
		// cross-check the amount
		updated_uacc, _ := MockTsStore.GetAccount(ctx, user_acc.Id)
		updated_bacc, _ := MockTsStore.GetAccount(ctx, brand_acc.Id)
		if updated_bacc.Amount != brand_last_amount-amount {
			// debitted from brand so less by 1000
			t.Fail()
		}
		if updated_uacc.Amount != user_last_amount+amount {
			// creditted to user so greater by 1000
			t.Fail()
		}
		user_last_amount = updated_uacc.Amount
		brand_last_amount = updated_bacc.Amount
	})
	t.Run("invalid payout with invalid ToId", func(t *testing.T) {
		invoice := Transaction{
			Id:     uuid.New().String(),
			FromId: brand_acc.Id,
			ToId:   uuid.New().String(),
			Amount: amount,
			Type:   "payout",
		}
		err := MockTsStore.Payout(ctx, &invoice)
		if err == nil {
			t.Fail()
			return
		}
	})
	t.Run("invalid payout with invalid FromId", func(t *testing.T) {
		invoice := Transaction{
			Id:     uuid.New().String(),
			FromId: uuid.New().String(),
			ToId:   user_acc.Id,
			Amount: amount,
			Type:   "payout",
		}
		err := MockTsStore.Payout(ctx, &invoice)
		if err == nil {
			t.Fail()
			return
		}
	})
	t.Run("making concurrent transactions", func(t *testing.T) {
		errs := make(chan error, 5)
		results := make(chan *Transaction, 5)

		var wg sync.WaitGroup // tick for every go routine started

		defer destroyAllTransactions()
		go func() {
			wg.Wait()
			close(errs)
			close(results)
		}()
		for range 5 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				invoice := Transaction{
					Id:     uuid.New().String(),
					FromId: brand_acc.Id,
					ToId:   user_acc.Id,
					Amount: amount,
					Type:   "payout",
				}
				err := MockTsStore.Payout(ctx, &invoice)
				if err != nil {
					log.Println(err.Error())
					t.Fail()
				}

				errs <- err
				results <- &invoice
			}()
		}
		// cross-check the amount
		for err := range errs {
			res := <-results
			if err != nil || res.Status == FailedTxStatus {
				t.Fail()
			}
			updated_uacc, _ := MockTsStore.GetAccount(ctx, user_acc.Id)
			updated_bacc, _ := MockTsStore.GetAccount(ctx, brand_acc.Id)
			u_diff := updated_uacc.Amount - user_last_amount
			b_diff := brand_last_amount - updated_bacc.Amount
			if b_diff != res.Amount {
				// debitted from brand so less by 6000
				log.Printf("brand diff error: got : %f\n", b_diff)
				t.Fail()
			}
			if u_diff != res.Amount {
				// creditted to user so greater by 6000
				log.Printf("user diff error: got : %f\n", u_diff)
				t.Fail()
			}
			user_last_amount = updated_uacc.Amount
			brand_last_amount = updated_bacc.Amount
		}
	})
}

func TestConcurrentTx(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	const amount float64 = 1000.0
	// create mock creator
	uid := uuid.New().String()
	generateCreator(ctx, uid)
	defer destroyCreator(ctx, uid)

	// create a mock brand
	bid := uuid.New().String()
	generateBrand(bid)
	defer destroyBrand(bid)

	// Create their accounts
	user_acc := generateAccounts(ctx, uid, "user")
	brand_acc := generateAccounts(ctx, bid, "brand")
	defer destroyAccounts(ctx, user_acc.Id, brand_acc.Id)

	// track the updated balances
	user_last_amount := user_acc.Amount
	brand_last_amount := brand_acc.Amount
	t.Run("concurrent deposit and withdraw", func(t *testing.T) {
		errs := make(chan error, 5)
		results := make(chan *Transaction, 5)

		var wg sync.WaitGroup // tick for every go routine started

		defer destroyAllTransactions()
		go func() {
			wg.Wait()
			close(errs)
			close(results)
		}()
		for i := range 3 {
			wg.Add(1)
			if i&1 == 0 {
				go func() {
					defer wg.Done()
					invoice := Transaction{
						Id:     uuid.New().String(),
						FromId: brand_acc.Id,
						ToId:   user_acc.Id,
						Amount: amount,
						Type:   "payout",
					}
					err := MockTsStore.Payout(ctx, &invoice)
					if err != nil {
						log.Println(err.Error())
						t.Fail()
					}

					errs <- err
					results <- &invoice
				}()
			} else {
				go func() {
					defer wg.Done()
					invoice := Transaction{
						Id:     uuid.New().String(),
						FromId: user_acc.Id,
						ToId:   brand_acc.Id,
						Amount: amount,
						Type:   "payout",
					}
					err := MockTsStore.Payout(ctx, &invoice)
					if err != nil {
						log.Println(err.Error())
						t.Fail()
					}

					errs <- err
					results <- &invoice
				}()
			}
		}
		// cross-check the amount
		for err := range errs {
			res := <-results
			if err != nil || res.Status == FailedTxStatus {
				t.Fail()
			}
			updated_uacc, _ := MockTsStore.GetAccount(ctx, user_acc.Id)
			updated_bacc, _ := MockTsStore.GetAccount(ctx, brand_acc.Id)
			u_diff := math.Abs(updated_uacc.Amount - user_last_amount)
			b_diff := math.Abs(brand_last_amount - updated_bacc.Amount)
			if b_diff != res.Amount {
				// debitted from brand so less by 6000
				log.Printf("brand diff error: got : %f\n", b_diff)
				t.Fail()
			}
			if u_diff != res.Amount {
				// creditted to user so greater by 6000
				log.Printf("user diff error: got : %f\n", u_diff)
				t.Fail()
			}
			user_last_amount = updated_uacc.Amount
			brand_last_amount = updated_bacc.Amount
		}
	})
}

func TestDeposit(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	const amount float64 = 1000.0
	// create mock creator
	uid := uuid.New().String()
	generateCreator(ctx, uid)
	defer destroyCreator(ctx, uid)

	// Create their accounts
	user_acc := generateAccounts(ctx, uid, "user")
	prevAmt := user_acc.Amount
	defer destroyAccounts(ctx, user_acc.Id)
	t.Run("mock deposit", func(t *testing.T) {
		defer destroyAllTransactions()
		invoice := Transaction{
			Id:     uuid.New().String(),
			FromId: user_acc.Id,
			ToId:   user_acc.Id,
			Amount: amount,
			Type:   "deposit",
		}
		err := MockTsStore.Deposit(ctx, &invoice)
		if err != nil {
			log.Println(err.Error())
			t.Fail()
		}
		updated_uacc, _ := MockTsStore.GetAccount(ctx, user_acc.Id)
		newAmt := updated_uacc.Amount
		if newAmt <= prevAmt {
			t.Fail()
		}
		if newAmt-prevAmt != amount {
			t.Fail()
		}
	})
}

func TestWithdraw(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	const amount float64 = 1000.0
	// create mock creator
	uid := uuid.New().String()
	generateCreator(ctx, uid)
	defer destroyCreator(ctx, uid)

	// Create their accounts
	user_acc := generateAccounts(ctx, uid, "user")
	prevAmt := user_acc.Amount
	defer destroyAccounts(ctx, user_acc.Id)
	defer destroyAllTransactions()
	t.Run("mock withdraw", func(t *testing.T) {
		invoice := Transaction{
			Id:     uuid.New().String(),
			FromId: user_acc.Id,
			ToId:   user_acc.Id,
			Amount: amount,
			Type:   "withdraw",
		}
		err := MockTsStore.Withdraw(ctx, &invoice)
		if err != nil {
			log.Println(err.Error())
			t.Fail()
		}
		updated_uacc, _ := MockTsStore.GetAccount(ctx, user_acc.Id)
		newAmt := updated_uacc.Amount
		if newAmt >= prevAmt {
			t.Fail()
		}
		if prevAmt-newAmt != amount {
			t.Fail()
		}
		prevAmt = newAmt
	})
	t.Run("withdraw bounce", func(t *testing.T) {
		invoice := Transaction{
			Id:     uuid.New().String(),
			FromId: user_acc.Id,
			ToId:   user_acc.Id,
			Amount: prevAmt + amount,
			Type:   "withdraw",
		}
		err := MockTsStore.Withdraw(ctx, &invoice)
		if err == nil {
			log.Println("withdraw bounced")
			t.Fail()
		}
	})
}

func TestDisableAccount(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	// create mock creator
	uid := uuid.New().String()
	generateCreator(ctx, uid)
	defer destroyCreator(ctx, uid)

	// Create their accounts
	user_acc := generateAccounts(ctx, uid, "user")
	defer destroyAccounts(ctx, user_acc.Id)
	t.Run("disabling an account", func(t *testing.T) {
		err := MockTsStore.DisableAccount(ctx, user_acc.Id)
		if err != nil {
			t.Fail()
		}
	})
}
