package main

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrNoRowsDeleted = errors.New("parcelStore: no rows has been deleted")
	ErrNoRowsUpdated = errors.New("parcelStore: no rows has benn updated")
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(ctx context.Context, p Parcel) (int, error) {
	query := "INSERT INTO parcel (client, status, address, created_at) VALUES (:client, :status, :address, :created_at)"

	res, err := s.db.ExecContext(ctx, query,
		sql.Named("client", p.Client),
		sql.Named("status", ParcelStatusRegistered),
		sql.Named("address", p.Address),
		sql.Named("created_at", p.CreatedAt))
	if err != nil {
		return 0, err
	}

	newId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(newId), nil
}

func (s ParcelStore) Get(ctx context.Context, number int) (Parcel, error) {
	query := "SELECT number, client, status, address, created_at FROM parcel WHERE number=:number"

	row := s.db.QueryRowContext(ctx, query, sql.Named("number", number))

	p := Parcel{}
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)

	if err != nil {
		return Parcel{}, err
	}
	return p, nil
}

func (s ParcelStore) GetByClient(ctx context.Context, client int) ([]Parcel, error) {

	query := "SELECT number, client, status, address, created_at FROM parcel WHERE client=:client"

	rows, err := s.db.QueryContext(ctx, query, sql.Named("client", client))
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	var res []Parcel
	res = make([]Parcel, 0)

	//итерируемся по строкам, так как между первым и вторым запросом могло пройти время
	for rows.Next() {
		var p Parcel
		err = rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (s ParcelStore) SetStatus(ctx context.Context, number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	query := "UPDATE parcel SET status = :status WHERE number = :number"

	res, err := s.db.ExecContext(ctx, query,
		sql.Named("number", number),
		sql.Named("status", status))

	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return ErrNoRowsUpdated
	}

	return nil
}

func (s ParcelStore) SetAddress(ctx context.Context, number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered
	query := "UPDATE parcel SET address = :address WHERE number = :number AND status = :status"

	res, err := s.db.ExecContext(ctx, query,
		sql.Named("address", address),
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered))

	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return ErrNoRowsUpdated
	}

	return nil
}

func (s ParcelStore) Delete(ctx context.Context, number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered
	query := "DELETE FROM parcel WHERE number = :number AND status = :status"

	res, err := s.db.ExecContext(ctx, query,
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered))

	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return ErrNoRowsDeleted
	}

	return nil
}
