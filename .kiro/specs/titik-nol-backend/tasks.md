# Rencana Implementasi: Titik Nol Backend (UC-03 — UC-15)

## Overview

Implementasi fitur-fitur backend Titik Nol yang belum dibangun, mengikuti arsitektur Clean Architecture yang sudah ada. Urutan implementasi: Domain Layer → Repository Layer → Usecase Layer → Delivery/HTTP Layer → Wiring di main.go. Setiap layer dibangun secara inkremental dan divalidasi dengan test. Semua handler menggunakan package `response` (RFC 7807), logging menggunakan `slog` dengan context, dan test menggunakan `testify` dengan naming `Test[Component]_[Scenario]`.

## Tasks

- [x] 1. Domain Layer — Entitas, Interface, dan Error
  - [x] 1.1 Buat file `internal/domain/account.go` dengan struct `Account`, konstanta `AccountType`, interface `AccountRepository`, interface `AccountUsecase`, dan DTO (`CreateAccountRequest`, `UpdateAccountRequest`)
    - Definisikan struct `Account` dengan GORM tags sesuai design (uuid PK, soft delete via `*time.Time`)
    - Definisikan `AccountType` (CASH, BANK, E_WALLET, CREDIT_CARD)
    - Definisikan interface `AccountRepository` dengan method: `WithTx`, `Create`, `Update`, `SoftDelete`, `GetByID`, `FetchByUserID`, `UpdateBalance`, `GetAllActive`
    - Definisikan interface `AccountUsecase` dengan method: `Create`, `Update`, `SoftDelete`, `FetchByUserID`
    - Definisikan DTO `CreateAccountRequest` dan `UpdateAccountRequest` dengan Gin binding tags
    - _Requirements: 2.1, 3.1, 3.3, 4.1, 5.1_

  - [x] 1.2 Buat file `internal/domain/transaction.go` dengan struct `Transaction`, konstanta `TransactionType`, interface `TransactionRepository`, interface `TransactionUsecase`, DTO, dan helper `CalculateBalanceDelta`
    - Definisikan struct `Transaction` dengan GORM tags (uuid PK, nullable `CategoryID` via `*uuid.UUID`, soft delete)
    - Definisikan `TransactionType` (INCOME, EXPENSE, TRANSFER, ADJUSTMENT)
    - Definisikan `TransactionQueryParams` untuk filter dan paginasi
    - Definisikan interface `TransactionRepository` dengan method: `WithTx`, `Create`, `Update`, `SoftDelete`, `GetByID`, `Fetch`, `SumByAccount`, `FetchRecent`
    - Definisikan interface `TransactionUsecase` dengan method: `Create`, `Update`, `SoftDelete`, `Fetch`
    - Definisikan DTO: `CreateTransactionRequest`, `CreateTransactionResponse`, `UpdateTransactionRequest`, `UpdateTransactionResponse`
    - Implementasikan fungsi `CalculateBalanceDelta(txType, amount) int64` sesuai tabel delta di design
    - _Requirements: 6.1, 6.2, 6.3, 7.1, 8.1, 9.1_

  - [x] 1.3 Buat file `internal/domain/category.go` dengan struct `Category`, konstanta `CategoryType`, interface `CategoryRepository`, interface `CategoryUsecase`, dan DTO
    - Definisikan struct `Category` dengan GORM tags (uuid PK, tanpa soft delete)
    - Definisikan `CategoryType` (INCOME, EXPENSE)
    - Definisikan interface `CategoryRepository` dengan method: `WithTx`, `Create`, `FetchByUserID`, `GetByID`, `CountByUserID`
    - Definisikan interface `CategoryUsecase` dengan method: `BulkCreate`, `FetchByUserID`
    - Definisikan DTO: `BulkCreateCategoryItem`, `BulkCreateCategoryRequest`
    - _Requirements: 11.1, 11.2, 12.1_

  - [x] 1.4 Buat file `internal/domain/onboarding.go` dengan interface `OnboardingUsecase` dan DTO
    - Definisikan interface `OnboardingUsecase` dengan method: `SetupAccounts`
    - Definisikan DTO: `SetupAccountItem`, `SetupAccountsRequest`, `SetupAccountsResponse`
    - _Requirements: 1.1, 1.7_

  - [x] 1.5 Buat file `internal/domain/dashboard.go` dengan interface `DashboardUsecase` dan DTO `DashboardSummary`
    - Definisikan interface `DashboardUsecase` dengan method: `GetSummary`
    - Definisikan struct `DashboardSummary` (TotalBalance, RecentTransactions, NeedsPaydaySetup)
    - _Requirements: 10.1, 10.2, 10.3_

  - [x] 1.6 Tambahkan domain error baru di `internal/domain/errors.go`
    - Tambahkan: `ErrAccountNotFound`, `ErrTransactionNotFound`, `ErrCategoryNotFound`, `ErrForbidden`, `ErrInvalidAccountType`, `ErrInvalidTxType`, `ErrInvalidCategoryType`, `ErrNegativeBalance`, `ErrEmptyBulkRequest`, `ErrAlreadyDeleted`, `ErrValidationFailed`
    - _Requirements: 1.4, 3.5, 4.3, 5.3, 6.10, 8.4, 9.6_


  - [x] 1.7 Tulis unit test untuk `CalculateBalanceDelta` di `internal/domain/transaction_test.go`
    - Test bahwa INCOME dan ADJUSTMENT menghasilkan delta positif (+amount)
    - Test bahwa EXPENSE menghasilkan delta negatif (-amount)
    - Test bahwa tipe tidak dikenal menghasilkan delta 0
    - Gunakan naming `TestCalculateBalanceDelta_Income`, `TestCalculateBalanceDelta_Expense`, dst.
    - _Requirements: 6.1, 6.2, 6.3, 9.2, 9.3, 9.4_

- [x] 2. Checkpoint — Validasi Domain Layer
  - Pastikan semua test lulus dengan `make test`, tanyakan ke user jika ada pertanyaan.

- [x] 3. Repository Layer — Implementasi Data Access
  - [x] 3.1 Buat file `internal/repository/account_repository.go` yang mengimplementasikan `domain.AccountRepository`
    - Implementasikan struct `accountRepository` dengan field `db *gorm.DB`
    - Implementasikan `WithTx(tx *gorm.DB) AccountRepository` yang mengembalikan instance baru dengan tx
    - Implementasikan `Create`: insert Account baru via GORM
    - Implementasikan `Update`: update field Account via GORM
    - Implementasikan `SoftDelete`: set `deleted_at` dengan `time.Now()`, filter `deleted_at IS NULL` dan `user_id`
    - Implementasikan `GetByID`: query by `id` dan `user_id`, filter `deleted_at IS NULL`
    - Implementasikan `FetchByUserID`: query semua Account milik user, filter `deleted_at IS NULL`, order `created_at DESC`
    - Implementasikan `UpdateBalance`: gunakan `gorm.Expr("balance + ?", delta)`, filter `deleted_at IS NULL`
    - Implementasikan `GetAllActive`: query semua Account dengan `deleted_at IS NULL` (untuk reconciliation)
    - Buat constructor `NewAccountRepository(db *gorm.DB) domain.AccountRepository`
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 3.1, 4.1, 4.2, 5.1, 5.4, 5.5_

  - [x] 3.2 Buat file `internal/repository/transaction_repository.go` yang mengimplementasikan `domain.TransactionRepository`
    - Implementasikan struct `transactionRepository` dengan field `db *gorm.DB`
    - Implementasikan `WithTx(tx *gorm.DB) TransactionRepository`
    - Implementasikan `Create`: insert Transaction baru
    - Implementasikan `Update`: update field Transaction yang diizinkan
    - Implementasikan `SoftDelete`: set `deleted_at`, filter `deleted_at IS NULL` dan `user_id`
    - Implementasikan `GetByID`: query by `id` dan `user_id`, filter `deleted_at IS NULL`
    - Implementasikan `Fetch`: query dengan filter opsional (`account_id`, `transaction_type`), paginasi (`page`, `per_page`), order `transaction_date DESC`, return total count
    - Implementasikan `SumByAccount`: SQL query COALESCE + CASE WHEN sesuai design untuk reconciliation
    - Implementasikan `FetchRecent`: query N transaksi terbaru milik user, filter `deleted_at IS NULL`, order `transaction_date DESC`
    - Buat constructor `NewTransactionRepository(db *gorm.DB) domain.TransactionRepository`
    - _Requirements: 6.1, 6.12, 7.1, 7.2, 7.3, 7.5, 7.6, 8.1, 8.7, 9.1, 9.7, 13.1, 13.4_

  - [x] 3.3 Buat file `internal/repository/category_repository.go` yang mengimplementasikan `domain.CategoryRepository`
    - Implementasikan struct `categoryRepository` dengan field `db *gorm.DB`
    - Implementasikan `WithTx(tx *gorm.DB) CategoryRepository`
    - Implementasikan `Create`: insert Category baru
    - Implementasikan `FetchByUserID`: query semua Category milik user, filter opsional by `type`, order `created_at DESC`
    - Implementasikan `GetByID`: query by `id` dan `user_id`
    - Implementasikan `CountByUserID`: count Category milik user (untuk dashboard flag)
    - Buat constructor `NewCategoryRepository(db *gorm.DB) domain.CategoryRepository`
    - _Requirements: 11.1, 12.1, 12.2, 12.3, 12.4, 10.3_

  - [x] 3.4 Buat mock untuk repository interfaces di `internal/domain/mocks/`
    - Buat `mock_account_repository.go` dengan mock `AccountRepository` menggunakan testify mock
    - Buat `mock_transaction_repository.go` dengan mock `TransactionRepository` menggunakan testify mock
    - Buat `mock_category_repository.go` dengan mock `CategoryRepository` menggunakan testify mock
    - _Requirements: semua — dibutuhkan untuk unit test usecase_

- [x] 4. Checkpoint — Validasi Repository Layer
  - Pastikan semua test lulus dengan `make test`, tanyakan ke user jika ada pertanyaan.

- [ ] 5. Usecase Layer — Implementasi Business Logic
  - [x] 5.1 Buat file `internal/usecase/account_usecase.go` yang mengimplementasikan `domain.AccountUsecase`
    - Struct `accountUsecase` dengan dependency: `accRepo domain.AccountRepository`, `txRepo domain.TransactionRepository`, `db *gorm.DB`
    - Implementasikan `Create`: validasi input → buat Account dalam db transaction → jika `InitialBalance > 0`, buat Transaction ADJUSTMENT dan set balance → return Account
    - Implementasikan `Update`: validasi ownership via `GetByID` → update nama → return Account
    - Implementasikan `SoftDelete`: validasi ownership via `GetByID` → soft delete
    - Implementasikan `FetchByUserID`: delegasi ke repository
    - Gunakan `slog.InfoContext`/`slog.ErrorContext` untuk logging
    - Buat constructor `NewAccountUsecase(accRepo, txRepo, db) domain.AccountUsecase`
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 3.1, 3.2, 3.3, 3.4, 3.5, 4.1, 4.2, 4.3, 4.4, 4.5, 5.1, 5.2, 5.3, 5.4, 5.5_

  - [x] 5.2 Tulis unit test untuk `AccountUsecase` di `internal/usecase/account_usecase_test.go`
    - Test `TestAccountUsecase_Create_Success`: buat akun dengan saldo awal > 0, verifikasi Account dan ADJUSTMENT dibuat
    - Test `TestAccountUsecase_Create_ZeroBalance`: buat akun dengan saldo 0, verifikasi tidak ada ADJUSTMENT
    - Test `TestAccountUsecase_Update_Success`: update nama akun
    - Test `TestAccountUsecase_Update_NotFound`: update akun yang tidak ada, expect error
    - Test `TestAccountUsecase_SoftDelete_Success`: soft delete akun
    - Test `TestAccountUsecase_FetchByUserID_Success`: fetch daftar akun
    - Gunakan mock repository dari task 3.4
    - _Requirements: 2.1, 3.1, 3.2, 4.1, 4.3, 5.1_

  - [x] 5.3 Buat file `internal/usecase/transaction_usecase.go` yang mengimplementasikan `domain.TransactionUsecase`
    - Struct `transactionUsecase` dengan dependency: `txRepo domain.TransactionRepository`, `accRepo domain.AccountRepository`, `catRepo domain.CategoryRepository`, `db *gorm.DB`
    - Implementasikan `Create`: dalam db transaction → validasi account ownership → validasi category ownership (jika ada) → buat Transaction → hitung delta via `CalculateBalanceDelta` → `UpdateBalance` → return response dengan balance baru
    - Implementasikan `Update`: dalam db transaction → validasi ownership → hitung selisih delta (oldDelta vs newDelta) → update Transaction → `UpdateBalance` dengan adjustment delta → return response
    - Implementasikan `SoftDelete`: dalam db transaction → validasi ownership → hitung reversal delta → soft delete Transaction → `UpdateBalance` dengan reversal → return
    - Implementasikan `Fetch`: validasi params → delegasi ke repository → return `PaginatedResult` dengan metadata
    - Gunakan `slog.InfoContext`/`slog.ErrorContext` untuk logging
    - Buat constructor `NewTransactionUsecase(txRepo, accRepo, catRepo, db) domain.TransactionUsecase`
    - _Requirements: 6.1–6.12, 7.1–7.7, 8.1–8.8, 9.1–9.7_

  - [x] 5.4 Tulis unit test untuk `TransactionUsecase` di `internal/usecase/transaction_usecase_test.go`
    - Test `TestTransactionUsecase_Create_Income`: buat transaksi INCOME, verifikasi balance bertambah
    - Test `TestTransactionUsecase_Create_Expense`: buat transaksi EXPENSE, verifikasi balance berkurang
    - Test `TestTransactionUsecase_Create_AccountNotFound`: account tidak ada, expect error
    - Test `TestTransactionUsecase_Create_WithCategory`: transaksi dengan category_id valid
    - Test `TestTransactionUsecase_Update_AmountChanged`: update amount, verifikasi selisih delta diterapkan
    - Test `TestTransactionUsecase_SoftDelete_Reversal`: soft delete, verifikasi reversal balance
    - Test `TestTransactionUsecase_Fetch_WithPagination`: fetch dengan paginasi
    - Gunakan mock repository dari task 3.4
    - _Requirements: 6.1, 6.2, 6.5, 6.9, 8.2, 9.2, 9.3, 7.1_

  - [x] 5.5 Buat file `internal/usecase/onboarding_usecase.go` yang mengimplementasikan `domain.OnboardingUsecase`
    - Struct `onboardingUsecase` dengan dependency: `accRepo domain.AccountRepository`, `txRepo domain.TransactionRepository`, `db *gorm.DB`
    - Implementasikan `SetupAccounts`: validasi semua item → dalam db transaction → loop: buat Account → jika `InitialBalance > 0`, buat Transaction ADJUSTMENT → kumpulkan hasil → return response
    - Validasi: nama tidak kosong, tipe valid, saldo tidak negatif, daftar tidak kosong
    - Jika validasi gagal, return error dengan indeks item yang gagal
    - Gunakan `slog.InfoContext` untuk logging
    - Buat constructor `NewOnboardingUsecase(accRepo, txRepo, db) domain.OnboardingUsecase`
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7_

  - [x] 5.6 Tulis unit test untuk `OnboardingUsecase` di `internal/usecase/onboarding_usecase_test.go`
    - Test `TestOnboardingUsecase_SetupAccounts_Success`: bulk insert 3 akun, 2 dengan saldo > 0
    - Test `TestOnboardingUsecase_SetupAccounts_ZeroBalance`: akun dengan saldo 0 tidak membuat ADJUSTMENT
    - Test `TestOnboardingUsecase_SetupAccounts_EmptyList`: daftar kosong, expect error
    - Test `TestOnboardingUsecase_SetupAccounts_InvalidItem`: item dengan nama kosong, expect error dengan indeks
    - Gunakan mock repository dari task 3.4
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6_

  - [x] 5.7 Buat file `internal/usecase/dashboard_usecase.go` yang mengimplementasikan `domain.DashboardUsecase`
    - Struct `dashboardUsecase` dengan dependency: `accRepo domain.AccountRepository`, `txRepo domain.TransactionRepository`, `catRepo domain.CategoryRepository`
    - Implementasikan `GetSummary`: fetch semua akun aktif user → hitung total balance → fetch 5 transaksi terbaru → cek count kategori → return `DashboardSummary`
    - Gunakan `slog.InfoContext` untuk logging
    - Buat constructor `NewDashboardUsecase(accRepo, txRepo, catRepo) domain.DashboardUsecase`
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5, 10.6_

  - [x] 5.8 Tulis unit test untuk `DashboardUsecase` di `internal/usecase/dashboard_usecase_test.go`
    - Test `TestDashboardUsecase_GetSummary_WithCategories`: user punya kategori, `needs_payday_setup` = false
    - Test `TestDashboardUsecase_GetSummary_NoCategories`: user tanpa kategori, `needs_payday_setup` = true
    - Test `TestDashboardUsecase_GetSummary_TotalBalance`: verifikasi total balance dari multiple akun
    - Gunakan mock repository dari task 3.4
    - _Requirements: 10.1, 10.2, 10.3, 10.4_

  - [x] 5.9 Buat file `internal/usecase/category_usecase.go` yang mengimplementasikan `domain.CategoryUsecase`
    - Struct `categoryUsecase` dengan dependency: `catRepo domain.CategoryRepository`, `db *gorm.DB`
    - Implementasikan `BulkCreate`: validasi semua item → dalam db transaction → loop: buat Category → return daftar Category
    - Validasi: nama tidak kosong, tipe valid (INCOME/EXPENSE), daftar tidak kosong
    - Implementasikan `FetchByUserID`: delegasi ke repository dengan filter opsional `type`
    - Gunakan `slog.InfoContext` untuk logging
    - Buat constructor `NewCategoryUsecase(catRepo, db) domain.CategoryUsecase`
    - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5, 11.6, 12.1, 12.2, 12.3, 12.4_

  - [x] 5.10 Tulis unit test untuk `CategoryUsecase` di `internal/usecase/category_usecase_test.go`
    - Test `TestCategoryUsecase_BulkCreate_Success`: bulk insert 3 kategori
    - Test `TestCategoryUsecase_BulkCreate_EmptyList`: daftar kosong, expect error
    - Test `TestCategoryUsecase_BulkCreate_InvalidItem`: item dengan nama kosong, expect error
    - Test `TestCategoryUsecase_FetchByUserID_WithFilter`: fetch dengan filter tipe
    - Gunakan mock repository dari task 3.4
    - _Requirements: 11.1, 11.2, 11.4, 11.5, 12.1, 12.3_

  - [x] 5.11 Buat file `internal/usecase/reconciliation_service.go` yang mengimplementasikan Reconciliation Service
    - Struct `reconciliationService` dengan dependency: `accRepo domain.AccountRepository`, `txRepo domain.TransactionRepository`
    - Implementasikan `ReconcileAll`: fetch semua akun aktif → untuk setiap akun, hitung expected balance via `SumByAccount` → bandingkan dengan stored balance → log WARN jika mismatch
    - Implementasikan `ReconcileAccount`: reconcile satu akun
    - Gunakan `slog.WarnContext` untuk log mismatch dengan detail `account_id`, `expected_balance`, `stored_balance`
    - TIDAK mengubah data balance (hanya reporting)
    - Buat constructor `NewReconciliationService(accRepo, txRepo) *reconciliationService`
    - _Requirements: 13.1, 13.2, 13.3, 13.4, 13.5, 13.6_

  - [x] 5.12 Tulis unit test untuk `ReconciliationService` di `internal/usecase/reconciliation_service_test.go`
    - Test `TestReconciliationService_ReconcileAccount_Match`: balance cocok, tidak ada warning
    - Test `TestReconciliationService_ReconcileAccount_Mismatch`: balance tidak cocok, verifikasi log warning
    - Gunakan mock repository dari task 3.4
    - _Requirements: 13.1, 13.2, 13.3_

- [x] 6. Checkpoint — Validasi Usecase Layer
  - Pastikan semua test lulus dengan `make test`, tanyakan ke user jika ada pertanyaan.


- [x] 7. Delivery/HTTP Layer — Handler dan Error Mapping
  - [x] 7.1 Buat helper `handleDomainError` di `internal/delivery/http/error_mapper.go`
    - Implementasikan fungsi `handleDomainError(c *gin.Context, err error)` yang memetakan domain error ke HTTP status menggunakan package `response`
    - Mapping: `ErrAccountNotFound`/`ErrTransactionNotFound`/`ErrCategoryNotFound`/`ErrAlreadyDeleted` → `response.NotFound`
    - Mapping: `ErrForbidden` → `response.NotFound` (404 bukan 403, sesuai requirement)
    - Mapping: `ErrValidationFailed`/`ErrNegativeBalance`/`ErrEmptyBulkRequest`/`ErrInvalidAccountType`/`ErrInvalidTxType`/`ErrInvalidCategoryType` → `response.BadRequest`
    - Default: `response.InternalServerError`
    - _Requirements: 1.4, 3.5, 4.3, 5.3, 6.10, 8.4, 9.6_

  - [x] 7.2 Buat file `internal/delivery/http/account_handler.go` yang mengimplementasikan AccountHandler
    - Struct `AccountHandler` dengan dependency: `accountUsecase domain.AccountUsecase`
    - Implementasikan `Create`: parse `CreateAccountRequest` via `ShouldBindJSON` → extract `user_id` dari context → panggil usecase → return `response.Success` (201)
    - Implementasikan `Fetch`: extract `user_id` → panggil usecase → return `response.Success` (200)
    - Implementasikan `Update`: parse UUID dari param `:id` → parse `UpdateAccountRequest` → extract `user_id` → panggil usecase → return `response.Success` (200)
    - Implementasikan `Delete`: parse UUID dari param `:id` → extract `user_id` → panggil usecase → return `response.Success` (200)
    - Gunakan `handleDomainError` untuk error handling
    - Buat constructor `NewAccountHandler(rg *gin.RouterGroup, uc domain.AccountUsecase)` yang mendaftarkan route: GET `/accounts`, POST `/accounts`, PUT `/accounts/:id`, DELETE `/accounts/:id`
    - _Requirements: 2.1, 2.4, 3.1, 3.5, 4.1, 4.3, 5.1, 5.3_

  - [x] 7.3 Tulis unit test untuk `AccountHandler` di `internal/delivery/http/account_handler_test.go`
    - Test `TestAccountHandler_Create_Success`: request valid, expect 201
    - Test `TestAccountHandler_Create_InvalidBody`: body tidak valid, expect 400
    - Test `TestAccountHandler_Fetch_Success`: fetch daftar akun, expect 200
    - Test `TestAccountHandler_Update_NotFound`: akun tidak ditemukan, expect 404
    - Test `TestAccountHandler_Delete_Success`: soft delete, expect 200
    - Gunakan mock usecase dan `httptest.NewRecorder`
    - _Requirements: 2.1, 3.1, 3.5, 4.3, 5.1_

  - [x] 7.4 Buat file `internal/delivery/http/transaction_handler.go` yang mengimplementasikan TransactionHandler
    - Struct `TransactionHandler` dengan dependency: `transactionUsecase domain.TransactionUsecase`
    - Implementasikan `Create`: parse `CreateTransactionRequest` → extract `user_id` → panggil usecase → return `response.Success` (201) dengan data transaksi dan balance baru
    - Implementasikan `Fetch`: extract `user_id` → parse query params (page, per_page, account_id, transaction_type) → panggil usecase → return `response.SuccessWithMeta` (200) dengan metadata paginasi
    - Implementasikan `Update`: parse UUID `:id` → parse `UpdateTransactionRequest` → extract `user_id` → panggil usecase → return `response.Success` (200)
    - Implementasikan `Delete`: parse UUID `:id` → extract `user_id` → panggil usecase → return `response.Success` (200)
    - Gunakan `handleDomainError` untuk error handling
    - Buat constructor `NewTransactionHandler(rg *gin.RouterGroup, uc domain.TransactionUsecase)` yang mendaftarkan route: POST `/transactions`, GET `/transactions`, PUT `/transactions/:id`, DELETE `/transactions/:id`
    - _Requirements: 6.1, 6.11, 7.1, 7.2, 7.4, 8.1, 8.8, 9.1_

  - [x] 7.5 Tulis unit test untuk `TransactionHandler` di `internal/delivery/http/transaction_handler_test.go`
    - Test `TestTransactionHandler_Create_Success`: request valid, expect 201
    - Test `TestTransactionHandler_Create_InvalidBody`: body tidak valid, expect 400
    - Test `TestTransactionHandler_Fetch_WithPagination`: fetch dengan query params, expect 200
    - Test `TestTransactionHandler_Update_Success`: update transaksi, expect 200
    - Test `TestTransactionHandler_Delete_NotFound`: transaksi tidak ditemukan, expect 404
    - Gunakan mock usecase dan `httptest.NewRecorder`
    - _Requirements: 6.1, 6.11, 7.1, 7.4, 8.1, 9.6_

  - [x] 7.6 Buat file `internal/delivery/http/onboarding_handler.go` yang mengimplementasikan OnboardingHandler
    - Struct `OnboardingHandler` dengan dependency: `onboardingUsecase domain.OnboardingUsecase`
    - Implementasikan `SetupAccounts`: parse `SetupAccountsRequest` → extract `user_id` → panggil usecase → return `response.Success` (201)
    - Gunakan `handleDomainError` untuk error handling
    - Buat constructor `NewOnboardingHandler(rg *gin.RouterGroup, uc domain.OnboardingUsecase)` yang mendaftarkan route: POST `/onboarding/accounts`
    - _Requirements: 1.1, 1.4, 1.7_

  - [x] 7.7 Tulis unit test untuk `OnboardingHandler` di `internal/delivery/http/onboarding_handler_test.go`
    - Test `TestOnboardingHandler_SetupAccounts_Success`: request valid, expect 201
    - Test `TestOnboardingHandler_SetupAccounts_InvalidBody`: body tidak valid, expect 400
    - Test `TestOnboardingHandler_SetupAccounts_EmptyList`: daftar kosong, expect 400
    - Gunakan mock usecase dan `httptest.NewRecorder`
    - _Requirements: 1.1, 1.4, 1.6_

  - [x] 7.8 Buat file `internal/delivery/http/dashboard_handler.go` yang mengimplementasikan DashboardHandler
    - Struct `DashboardHandler` dengan dependency: `dashboardUsecase domain.DashboardUsecase`
    - Implementasikan `GetSummary`: extract `user_id` → panggil usecase → return `response.Success` (200)
    - Gunakan `handleDomainError` untuk error handling
    - Buat constructor `NewDashboardHandler(rg *gin.RouterGroup, uc domain.DashboardUsecase)` yang mendaftarkan route: GET `/dashboard`
    - _Requirements: 10.1, 10.5_

  - [x] 7.9 Tulis unit test untuk `DashboardHandler` di `internal/delivery/http/dashboard_handler_test.go`
    - Test `TestDashboardHandler_GetSummary_Success`: request valid, expect 200
    - Gunakan mock usecase dan `httptest.NewRecorder`
    - _Requirements: 10.1_

  - [x] 7.10 Buat file `internal/delivery/http/category_handler.go` yang mengimplementasikan CategoryHandler
    - Struct `CategoryHandler` dengan dependency: `categoryUsecase domain.CategoryUsecase`
    - Implementasikan `BulkCreate`: parse `BulkCreateCategoryRequest` → extract `user_id` → panggil usecase → return `response.Success` (201)
    - Implementasikan `Fetch`: extract `user_id` → parse query param `type` (opsional) → panggil usecase → return `response.Success` (200)
    - Gunakan `handleDomainError` untuk error handling
    - Buat constructor `NewCategoryHandler(rg *gin.RouterGroup, uc domain.CategoryUsecase)` yang mendaftarkan route: POST `/categories`, GET `/categories`
    - _Requirements: 11.1, 11.6, 12.1, 12.3_

  - [x] 7.11 Tulis unit test untuk `CategoryHandler` di `internal/delivery/http/category_handler_test.go`
    - Test `TestCategoryHandler_BulkCreate_Success`: request valid, expect 201
    - Test `TestCategoryHandler_BulkCreate_InvalidBody`: body tidak valid, expect 400
    - Test `TestCategoryHandler_Fetch_Success`: fetch daftar kategori, expect 200
    - Test `TestCategoryHandler_Fetch_WithTypeFilter`: fetch dengan filter tipe, expect 200
    - Gunakan mock usecase dan `httptest.NewRecorder`
    - _Requirements: 11.1, 11.5, 12.1, 12.3_

  - [x] 7.12 Buat mock untuk usecase interfaces di `internal/domain/mocks/`
    - Buat `mock_account_usecase.go` dengan mock `AccountUsecase`
    - Buat `mock_transaction_usecase.go` dengan mock `TransactionUsecase`
    - Buat `mock_onboarding_usecase.go` dengan mock `OnboardingUsecase`
    - Buat `mock_dashboard_usecase.go` dengan mock `DashboardUsecase`
    - Buat `mock_category_usecase.go` dengan mock `CategoryUsecase`
    - _Requirements: semua — dibutuhkan untuk unit test handler_

- [x] 8. Checkpoint — Validasi Delivery Layer
  - Pastikan semua test lulus dengan `make test`, tanyakan ke user jika ada pertanyaan.

- [x] 9. Wiring — Integrasi di main.go
  - [x] 9.1 Update `cmd/api/main.go` untuk menginisialisasi dan menghubungkan semua komponen baru
    - Import package repository, usecase, dan delivery baru
    - Inisialisasi repository: `NewAccountRepository(db)`, `NewTransactionRepository(db)`, `NewCategoryRepository(db)`
    - Inisialisasi usecase: `NewAccountUsecase(accRepo, txRepo, db)`, `NewTransactionUsecase(txRepo, accRepo, catRepo, db)`, `NewOnboardingUsecase(accRepo, txRepo, db)`, `NewDashboardUsecase(accRepo, txRepo, catRepo)`, `NewCategoryUsecase(catRepo, db)`, `NewReconciliationService(accRepo, txRepo)`
    - Buat route group `v1 := r.Group("/api/v1")` dengan `authMiddleware`
    - Daftarkan handler: `NewAccountHandler(v1, accountUsecase)`, `NewTransactionHandler(v1, transactionUsecase)`, `NewOnboardingHandler(v1, onboardingUsecase)`, `NewDashboardHandler(v1, dashboardUsecase)`, `NewCategoryHandler(v1, categoryUsecase)`
    - Pastikan handler auth yang sudah ada tetap berfungsi
    - _Requirements: 1.1, 2.1, 3.1, 6.1, 7.1, 10.1, 11.1, 12.1_

- [x] 10. Checkpoint Final — Validasi Keseluruhan
  - Pastikan semua test lulus dengan `make test`, tanyakan ke user jika ada pertanyaan.

## Catatan

- Task dengan tanda `*` bersifat opsional dan dapat dilewati untuk MVP lebih cepat
- Setiap task mereferensikan requirement spesifik untuk traceability
- Checkpoint memastikan validasi inkremental di setiap layer
- Semua handler menggunakan package `response` (RFC 7807), bukan raw `gin.H{}`
- Logging menggunakan `slog` dengan context (`slog.InfoContext`, `slog.WarnContext`, `slog.ErrorContext`)
- Test menggunakan naming `Test[Component]_[Scenario]` tanpa `t.Run` subtests
