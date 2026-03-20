# Dokumen Requirements — Titik Nol Backend

## Pendahuluan

Dokumen ini mendefinisikan requirements untuk fitur-fitur backend Project Titik Nol yang belum diimplementasi (UC-03 sampai UC-15). Titik Nol adalah aplikasi manajemen keuangan pribadi yang membantu pengguna mencatat saldo awal (titik nol), mengelola akun sumber dana, mencatat transaksi harian, dan melihat ringkasan dashboard. Backend dibangun dengan Go (Gin + GORM + PostgreSQL) menggunakan arsitektur Clean Architecture. Modul Autentikasi (UC-01: Google Login/Register, UC-02: Get Current User Profile) sudah terimplementasi.

## Glosarium

- **API**: Application Programming Interface, antarmuka komunikasi antara client dan server
- **Sistem**: Backend API Titik Nol yang berjalan di server
- **Onboarding_Service**: Komponen usecase yang menangani proses setup awal akun pengguna (titik nol)
- **Account_Service**: Komponen usecase yang menangani operasi CRUD pada entitas akun sumber dana
- **Transaction_Service**: Komponen usecase yang menangani operasi CRUD pada entitas transaksi keuangan
- **Dashboard_Service**: Komponen usecase yang menangani agregasi data untuk ringkasan dashboard
- **Category_Service**: Komponen usecase yang menangani operasi pada entitas kategori transaksi
- **Reconciliation_Service**: Komponen background task yang memverifikasi konsistensi saldo akun
- **Account**: Entitas sumber dana milik pengguna (CASH, BANK, E_WALLET, CREDIT_CARD) dengan saldo dalam satuan terkecil (Rupiah, BIGINT)
- **Transaction**: Entitas catatan transaksi keuangan dengan tipe INCOME, EXPENSE, TRANSFER, atau ADJUSTMENT
- **Category**: Entitas kategori transaksi dengan tipe INCOME atau EXPENSE, digunakan pada fitur Payday Mode
- **Soft_Delete**: Mekanisme penghapusan logis menggunakan field `deleted_at` (TIMESTAMP), data tidak dihapus secara fisik dari database
- **ADJUSTMENT**: Tipe transaksi khusus yang digunakan saat onboarding untuk mencatat saldo awal sebagai audit trail
- **Authenticated_User**: Pengguna yang telah melewati auth middleware dan memiliki `user_id` valid di context request
- **BIGINT**: Tipe data integer 64-bit yang digunakan untuk menyimpan nilai moneter dalam satuan terkecil (Rupiah)
- **Pagination**: Mekanisme pembagian data menjadi halaman-halaman dengan parameter `page` dan `per_page`

## Requirements

### Requirement 1: Setup Akun Awal (Onboarding / Titik Nol) — UC-03

**User Story:** Sebagai pengguna baru, saya ingin memasukkan daftar akun beserta saldo awal saya secara bulk, sehingga saya memiliki titik nol keuangan yang tercatat dengan audit trail.

#### Acceptance Criteria

1. WHEN Authenticated_User mengirim request bulk insert berisi daftar akun (nama, tipe, saldo awal), THE Onboarding_Service SHALL membuat semua Account dalam satu database transaction
2. WHEN sebuah Account berhasil dibuat dengan saldo awal lebih dari 0, THE Onboarding_Service SHALL membuat satu Transaction bertipe ADJUSTMENT dengan amount sama dengan saldo awal untuk Account tersebut
3. WHEN sebuah Account dibuat dengan saldo awal sama dengan 0, THE Onboarding_Service SHALL membuat Account tanpa membuat Transaction ADJUSTMENT
4. IF salah satu Account dalam bulk insert gagal validasi, THEN THE Onboarding_Service SHALL membatalkan seluruh operasi (rollback) dan mengembalikan pesan error yang menjelaskan Account mana yang gagal
5. THE Onboarding_Service SHALL memvalidasi bahwa setiap Account dalam request memiliki nama yang tidak kosong, tipe yang valid (CASH, BANK, E_WALLET, CREDIT_CARD), dan saldo awal yang tidak negatif
6. IF Authenticated_User mengirim request dengan daftar akun kosong, THEN THE Onboarding_Service SHALL mengembalikan error validasi
7. WHEN bulk insert berhasil, THE Onboarding_Service SHALL mengembalikan daftar Account yang telah dibuat beserta Transaction ADJUSTMENT terkait

### Requirement 2: Daftar Akun Pengguna — UC-04

**User Story:** Sebagai pengguna, saya ingin melihat daftar semua akun sumber dana saya, sehingga saya dapat memantau saldo masing-masing akun.

#### Acceptance Criteria

1. WHEN Authenticated_User mengirim request untuk melihat daftar akun, THE Account_Service SHALL mengembalikan semua Account milik Authenticated_User yang belum di-soft-delete
2. THE Account_Service SHALL menyertakan saldo terkini setiap Account dalam response
3. THE Account_Service SHALL mengurutkan daftar Account berdasarkan waktu pembuatan (terbaru di atas)
4. THE Account_Service SHALL hanya mengembalikan Account milik Authenticated_User yang sedang login (isolasi data per pengguna)

### Requirement 3: Tambah Akun Baru — UC-05

**User Story:** Sebagai pengguna, saya ingin menambahkan akun sumber dana baru setelah onboarding, sehingga saya dapat mencatat sumber dana tambahan.

#### Acceptance Criteria

1. WHEN Authenticated_User mengirim request untuk membuat Account baru dengan nama, tipe, dan saldo awal, THE Account_Service SHALL membuat Account baru milik Authenticated_User
2. WHEN Account baru dibuat dengan saldo awal lebih dari 0, THE Account_Service SHALL membuat Transaction bertipe ADJUSTMENT sebagai pencatatan saldo awal
3. THE Account_Service SHALL memvalidasi bahwa nama Account tidak kosong dan tipe Account adalah salah satu dari CASH, BANK, E_WALLET, atau CREDIT_CARD
4. THE Account_Service SHALL memvalidasi bahwa saldo awal tidak bernilai negatif
5. IF validasi gagal, THEN THE Account_Service SHALL mengembalikan error dengan detail field yang tidak valid

### Requirement 4: Update Detail Akun — UC-06

**User Story:** Sebagai pengguna, saya ingin mengubah nama atau detail akun saya, sehingga informasi akun tetap akurat.

#### Acceptance Criteria

1. WHEN Authenticated_User mengirim request update untuk Account tertentu, THE Account_Service SHALL memperbarui field yang diizinkan (nama akun)
2. THE Account_Service SHALL memvalidasi bahwa Account yang di-update adalah milik Authenticated_User
3. IF Account tidak ditemukan atau bukan milik Authenticated_User, THEN THE Account_Service SHALL mengembalikan error 404
4. THE Account_Service SHALL tidak mengizinkan perubahan langsung pada field balance melalui endpoint update (balance hanya berubah melalui transaksi)
5. THE Account_Service SHALL memvalidasi bahwa nama Account yang baru tidak kosong

### Requirement 5: Soft Delete Akun — UC-07

**User Story:** Sebagai pengguna, saya ingin menghapus akun yang tidak lagi digunakan, sehingga daftar akun saya tetap rapi tanpa kehilangan data historis.

#### Acceptance Criteria

1. WHEN Authenticated_User mengirim request delete untuk Account tertentu, THE Account_Service SHALL melakukan Soft_Delete dengan mengisi field `deleted_at`
2. THE Account_Service SHALL memvalidasi bahwa Account yang di-delete adalah milik Authenticated_User
3. IF Account tidak ditemukan atau bukan milik Authenticated_User, THEN THE Account_Service SHALL mengembalikan error 404
4. WHEN Account berhasil di-soft-delete, THE Account_Service SHALL tidak menampilkan Account tersebut pada endpoint daftar akun
5. IF Account sudah dalam status soft-deleted, THEN THE Account_Service SHALL mengembalikan error 404

### Requirement 6: Buat Transaksi Quick-Log — UC-08

**User Story:** Sebagai pengguna, saya ingin mencatat transaksi pemasukan atau pengeluaran dengan cepat, sehingga saya dapat melacak arus keuangan harian saya secara real-time.

#### Acceptance Criteria

1. WHEN Authenticated_User mengirim request untuk membuat Transaction baru dengan tipe EXPENSE, THE Transaction_Service SHALL mengurangi balance Account terkait secara atomik sebesar amount transaksi
2. WHEN Authenticated_User mengirim request untuk membuat Transaction baru dengan tipe INCOME, THE Transaction_Service SHALL menambah balance Account terkait secara atomik sebesar amount transaksi
3. WHEN Authenticated_User mengirim request untuk membuat Transaction baru dengan tipe ADJUSTMENT, THE Transaction_Service SHALL menyesuaikan balance Account terkait secara atomik sebesar amount transaksi
4. THE Transaction_Service SHALL memvalidasi bahwa amount transaksi lebih dari 0
5. THE Transaction_Service SHALL memvalidasi bahwa Account yang direferensikan adalah milik Authenticated_User dan belum di-soft-delete
6. THE Transaction_Service SHALL memvalidasi bahwa tipe transaksi adalah salah satu dari INCOME, EXPENSE, atau ADJUSTMENT
7. THE Transaction_Service SHALL menyimpan field `transaction_date` yang dikirim oleh client
8. THE Transaction_Service SHALL mengizinkan field `category_id` bernilai null untuk mendukung pencatatan cepat tanpa kategori
9. IF `category_id` disertakan, THE Transaction_Service SHALL memvalidasi bahwa Category tersebut adalah milik Authenticated_User
10. IF Account tidak ditemukan atau bukan milik Authenticated_User, THEN THE Transaction_Service SHALL mengembalikan error 404
11. WHEN Transaction berhasil dibuat, THE Transaction_Service SHALL mengembalikan data Transaction beserta balance Account yang telah diperbarui
12. THE Transaction_Service SHALL menjalankan pembuatan Transaction dan update balance Account dalam satu database transaction untuk menjamin atomicity

### Requirement 7: Riwayat Transaksi (Paginasi) — UC-09

**User Story:** Sebagai pengguna, saya ingin melihat riwayat transaksi saya dengan paginasi, sehingga saya dapat menelusuri catatan keuangan secara efisien.

#### Acceptance Criteria

1. WHEN Authenticated_User mengirim request untuk melihat riwayat transaksi, THE Transaction_Service SHALL mengembalikan daftar Transaction milik Authenticated_User yang belum di-soft-delete
2. THE Transaction_Service SHALL mendukung paginasi dengan parameter `page` (default: 1) dan `per_page` (default: 20, maksimum: 100)
3. THE Transaction_Service SHALL mengurutkan Transaction berdasarkan `transaction_date` dari yang terbaru
4. THE Transaction_Service SHALL menyertakan metadata paginasi dalam response (total_items, total_pages, page, per_page)
5. WHERE parameter filter `account_id` disertakan, THE Transaction_Service SHALL hanya mengembalikan Transaction untuk Account tersebut
6. WHERE parameter filter `transaction_type` disertakan, THE Transaction_Service SHALL hanya mengembalikan Transaction dengan tipe yang sesuai
7. THE Transaction_Service SHALL hanya mengembalikan Transaction milik Authenticated_User yang sedang login (isolasi data per pengguna)

### Requirement 8: Update Transaksi — UC-10

**User Story:** Sebagai pengguna, saya ingin mengubah detail transaksi yang sudah dicatat, sehingga saya dapat memperbaiki kesalahan pencatatan.

#### Acceptance Criteria

1. WHEN Authenticated_User mengirim request update untuk Transaction tertentu, THE Transaction_Service SHALL memperbarui field yang diizinkan (amount, note, category_id, transaction_date)
2. WHEN amount Transaction diubah, THE Transaction_Service SHALL menghitung selisih antara amount lama dan amount baru, lalu menyesuaikan balance Account secara atomik
3. THE Transaction_Service SHALL memvalidasi bahwa Transaction yang di-update adalah milik Authenticated_User dan belum di-soft-delete
4. IF Transaction tidak ditemukan atau bukan milik Authenticated_User, THEN THE Transaction_Service SHALL mengembalikan error 404
5. THE Transaction_Service SHALL memvalidasi bahwa amount baru lebih dari 0
6. THE Transaction_Service SHALL tidak mengizinkan perubahan tipe transaksi (transaction_type) dan account_id
7. THE Transaction_Service SHALL menjalankan update Transaction dan penyesuaian balance Account dalam satu database transaction untuk menjamin atomicity
8. WHEN Transaction berhasil di-update, THE Transaction_Service SHALL mengembalikan data Transaction yang telah diperbarui beserta balance Account terkini

### Requirement 9: Soft Delete Transaksi — UC-11

**User Story:** Sebagai pengguna, saya ingin menghapus transaksi yang salah, sehingga saldo akun saya kembali akurat tanpa kehilangan jejak audit.

#### Acceptance Criteria

1. WHEN Authenticated_User mengirim request delete untuk Transaction tertentu, THE Transaction_Service SHALL melakukan Soft_Delete dengan mengisi field `deleted_at`
2. WHEN Transaction bertipe EXPENSE di-soft-delete, THE Transaction_Service SHALL menambah kembali balance Account sebesar amount transaksi (reversal)
3. WHEN Transaction bertipe INCOME di-soft-delete, THE Transaction_Service SHALL mengurangi balance Account sebesar amount transaksi (reversal)
4. WHEN Transaction bertipe ADJUSTMENT di-soft-delete, THE Transaction_Service SHALL membalikkan efek amount transaksi pada balance Account (reversal)
5. THE Transaction_Service SHALL memvalidasi bahwa Transaction yang di-delete adalah milik Authenticated_User dan belum di-soft-delete
6. IF Transaction tidak ditemukan atau bukan milik Authenticated_User, THEN THE Transaction_Service SHALL mengembalikan error 404
7. THE Transaction_Service SHALL menjalankan soft-delete Transaction dan reversal balance Account dalam satu database transaction untuk menjamin atomicity

### Requirement 10: Dashboard Summary — UC-12

**User Story:** Sebagai pengguna, saya ingin melihat ringkasan keuangan saya di dashboard, sehingga saya mendapat gambaran cepat tentang kondisi keuangan terkini.

#### Acceptance Criteria

1. WHEN Authenticated_User mengirim request untuk dashboard summary, THE Dashboard_Service SHALL mengembalikan total balance dari semua Account aktif (belum di-soft-delete) milik Authenticated_User
2. WHEN Authenticated_User mengirim request untuk dashboard summary, THE Dashboard_Service SHALL mengembalikan daftar 5 Transaction terbaru milik Authenticated_User yang belum di-soft-delete
3. WHEN Authenticated_User belum memiliki Category sama sekali, THE Dashboard_Service SHALL menyertakan flag `needs_payday_setup: true` dalam response
4. WHEN Authenticated_User sudah memiliki minimal satu Category, THE Dashboard_Service SHALL menyertakan flag `needs_payday_setup: false` dalam response
5. THE Dashboard_Service SHALL hanya menggunakan data milik Authenticated_User yang sedang login (isolasi data per pengguna)
6. THE Dashboard_Service SHALL mengembalikan total balance dalam satuan BIGINT (Rupiah, satuan terkecil)

### Requirement 11: Setup Kategori (Bulk Insert) — UC-13

**User Story:** Sebagai pengguna, saya ingin membuat daftar kategori transaksi secara bulk untuk mempersiapkan fitur Payday Mode, sehingga saya dapat mengkategorikan pengeluaran dan pemasukan.

#### Acceptance Criteria

1. WHEN Authenticated_User mengirim request bulk insert berisi daftar kategori (nama, tipe, icon), THE Category_Service SHALL membuat semua Category dalam satu database transaction
2. THE Category_Service SHALL memvalidasi bahwa setiap Category memiliki nama yang tidak kosong dan tipe yang valid (INCOME atau EXPENSE)
3. THE Category_Service SHALL mengizinkan field `icon` bernilai kosong (opsional)
4. IF salah satu Category dalam bulk insert gagal validasi, THEN THE Category_Service SHALL membatalkan seluruh operasi (rollback) dan mengembalikan pesan error yang menjelaskan Category mana yang gagal
5. IF Authenticated_User mengirim request dengan daftar kategori kosong, THEN THE Category_Service SHALL mengembalikan error validasi
6. WHEN bulk insert berhasil, THE Category_Service SHALL mengembalikan daftar Category yang telah dibuat

### Requirement 12: Daftar Kategori Pengguna — UC-14

**User Story:** Sebagai pengguna, saya ingin melihat daftar kategori transaksi saya, sehingga saya dapat memilih kategori saat mencatat transaksi.

#### Acceptance Criteria

1. WHEN Authenticated_User mengirim request untuk melihat daftar kategori, THE Category_Service SHALL mengembalikan semua Category milik Authenticated_User
2. THE Category_Service SHALL hanya mengembalikan Category milik Authenticated_User yang sedang login (isolasi data per pengguna)
3. WHERE parameter filter `type` disertakan (INCOME atau EXPENSE), THE Category_Service SHALL hanya mengembalikan Category dengan tipe yang sesuai
4. THE Category_Service SHALL mengurutkan daftar Category berdasarkan waktu pembuatan (terbaru di atas)

### Requirement 13: Balance Reconciliation (Background Task) — UC-15

**User Story:** Sebagai sistem, saya ingin memverifikasi konsistensi saldo akun secara berkala, sehingga ketidaksesuaian antara saldo tercatat dan total transaksi dapat terdeteksi.

#### Acceptance Criteria

1. WHEN Reconciliation_Service dijalankan untuk sebuah Account, THE Reconciliation_Service SHALL menghitung ulang expected balance berdasarkan penjumlahan semua Transaction aktif (belum di-soft-delete) milik Account tersebut
2. THE Reconciliation_Service SHALL membandingkan expected balance dengan balance yang tersimpan di field `balance` pada Account
3. IF expected balance tidak sama dengan stored balance, THEN THE Reconciliation_Service SHALL mencatat ketidaksesuaian tersebut ke dalam log dengan level WARN beserta detail account_id, expected balance, dan stored balance
4. THE Reconciliation_Service SHALL menghitung expected balance dengan menjumlahkan amount Transaction bertipe INCOME dan ADJUSTMENT, lalu mengurangi amount Transaction bertipe EXPENSE
5. THE Reconciliation_Service SHALL dapat dijalankan untuk semua Account aktif milik semua pengguna
6. THE Reconciliation_Service SHALL tidak mengubah data balance secara otomatis (hanya melaporkan ketidaksesuaian)
