include .env

# Command untuk menjalankan migrasi ke atas
migrate-up:
	@goose -dir migrations postgres "${DB_DSN}" up

# Command untuk membatalkan 1 migrasi terakhir
migrate-down:
	@goose -dir migrations postgres "${DB_DSN}" down

# Command untuk melihat status migrasi
migrate-status:
	@goose -dir migrations postgres "${DB_DSN}" status

# Membuat file migrasi baru (contoh: make migrate-create name=add_users_table)
migrate-create:
	@goose -dir migrations create $(name) sql

seed:
	go run cmd/seeder/main.go

run-dev:
	@go run .\cmd\api
