## 1. Ringkasan Eksekutif

Sistem Operasional Kopi Nofu adalah platform digital terintegrasi yang dirancang untuk memfasilitasi penjualan kopi keliling (menggunakan rider) dan manajemen inventaris oleh admin. Sistem ini akan memastikan transparansi stok, mempermudah pelaporan penjualan secara real-time, mengelola komisi rider, dan mengefisiensikan proses pembukuan harian.

## 2. Arsitektur Teknis & Tech Stack

Sistem ini terbagi menjadi komponen utama untuk memastikan operasional yang lancar:

- **Aplikasi Rider:** Dibangun menggunakan React Native (untuk fleksibilitas cross-platform di perangkat mobile para rider).
- **Aplikasi Admin:** Dibangun menggunakan Flutter (sebagai dashboard operasional yang responsif dan cepat).
- **Sistem Backend:** API terpusat yang menghubungkan dan mengintegrasikan kedua aplikasi (Admin & Rider) serta mengelola database sentral. Integrasi ini menyediakan RESTful API/GraphQL untuk komunikasi yang mulus antara aplikasi Rider dan Admin.
- **Database:** Menggunakan **PostgreSQL** untuk penyimpanan data yang tangguh dan terstruktur dengan baik (sesuai penambahan requirement terbaru).

## 3. Matriks Fitur & Ruang Lingkup

Tabel berikut merangkum pembagian fitur antara Aplikasi Rider (dipegang oleh rider di lapangan ) dan Aplikasi Admin (digunakan oleh staf operasional di hub/warehouse ).

| **Modul Utama**            | **Fitur Aplikasi Rider (React Native)**                                                                                | **Fitur Aplikasi Admin (Flutter)**                                                                                             |
| -------------------------- | ---------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| **Inisiasi Stok Awal**     | Menerima notifikasi inisiasi stok harian dari Admin. Memberikan respons "Terima" atau "Tolak" saat inisiasi stok awal. | Mengalokasikan dan menginisiasi jumlah stok produk kepada masing-masing rider di awal shift.                                   |
| **Manajemen Penjualan**    | Melaporkan pengurangan stok saat terjadi transaksi dengan pembeli secara real-time.                                    | Melihat riwayat dan detail seluruh transaksi penjualan yang terjadi.                                                           |
| **Isu Stok & Cacat**       | Melaporkan stok produk yang cacat/rusak beserta alasannya.                                                             | Memantau laporan stok cacat yang dikirimkan oleh rider.                                                                        |
| **Restock (Isi Ulang)**    | Mengajukan permohonan penambahan stok ke Admin jika stok menipis.                                                      | Menerima dan menyetujui/menolak pengajuan restock dari rider.                                                                  |
| **Riwayat & Rekapitulasi** | Melihat rekap data penjualan berdasarkan filter: Harian, Mingguan, dan Bulanan.                                        | Melakukan rekap penjualan dan mencocokkan sisa stok fisik dengan sistem saat akhir shift. Menutup sesi operasional tiap rider. |
| **Komisi**                 | Melihat akumulasi komisi penjualan yang didapatkan.                                                                    | Sistem backend otomatis merekap data komisi berdasarkan transaksi.                                                             |
| **Manajemen Produk**       | (Tidak ada akses)                                                                                                      | Mengelola master data produk (tambah, edit, nonaktif).                                                                         |
| **Manajemen Pengguna**     | (Tidak ada akses)                                                                                                      | Mengelola data rider (tambah, edit, nonaktif akun rider).                                                                      |

> **Catatan Backend:** Backend bertugas menyimpan data pengguna (rider dan admin), master produk, catatan transaksi, riwayat stok harian, komisi, dan laporan produk cacat di dalam database PostgreSQL. Backend juga memastikan sinkronisasi real-time agar pengurangan stok di aplikasi rider langsung tercermin pada dashboard admin.

## 4. Alur Kerja Utama (User Flows)

### A. Alur Inisiasi Stok (Awal Shift)

1. Admin membuka aplikasi Admin dan melakukan input stok awal untuk Rider A.
2. Sistem Backend mengirimkan data ke aplikasi Rider A.
3. Rider A membuka aplikasinya, melihat rincian stok awal, dan menghitung fisik barang.
4. Rider A menekan "Terima" (stok resmi tercatat sebagai tanggung jawab rider) atau "Tolak" (admin perlu merevisi data).

### B. Alur Transaksi & Restock (Operasional Tengah Shift)

1. Rider melakukan transaksi dengan pembeli, lalu menekan tombol "Jual" di aplikasi. Stok otomatis berkurang di database.
2. Jika ada kopi yang tumpah/rusak, Rider membuat laporan "Stok Cacat". Stok berkurang, masuk kategori waste.
3. Jika stok sisa sedikit, Rider menekan tombol "Ajukan Restock". Admin melihat request tersebut, menyiapkan barang, dan mengonfirmasi penambahan stok.

### C. Alur Penutupan (Akhir Shift)

1. Rider kembali ke hub dan menyerahkan sisa stok fisik serta uang tunai/bukti transfer.
2. Admin membuka menu "Rekap Penutupan".
3. Admin memvalidasi total transaksi, laporan cacat, dan sisa stok fisik.
4. Admin menekan "Tutup Sesi". Sistem otomatis menghitung komisi final Rider untuk hari tersebut.

## 5. Pertimbangan Non-Fungsional

- **Keamanan Data:** Sistem akan menggunakan autentikasi berbasis token (JWT) untuk login Admin dan Rider.
- **Reliabilitas Jaringan:** Aplikasi Rider harus memiliki penanganan error yang baik jika koneksi internet di jalan tidak stabil (misal: local caching sementara untuk laporan transaksi sebelum di-push ke backend).
- **Antarmuka Pengguna (UI/UX) Rider:** Aplikasi Rider menggunakan interface dengan tombol besar dan flow yang cepat (mempertimbangkan kondisi rider yang bekerja di jalan/luar ruangan).
- **Antarmuka Pengguna (UI/UX) Admin:** Aplikasi Admin menggunakan desain tabel yang informatif (dashboard-style).
