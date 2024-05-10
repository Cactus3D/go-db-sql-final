package main

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const TestExecTimeout = 5 * time.Second

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	ctx, cancelfunc := context.WithTimeout(context.Background(), TestExecTimeout)
	defer cancelfunc()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	number, err := store.Add(ctx, parcel)
	require.NoError(t, err)
	require.Greater(t, number, 0)

	parcel.Number = number

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	newParcel, err := store.Get(ctx, number)
	require.NoError(t, err)
	assert.Equal(t, parcel, newParcel)

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(ctx, number)
	require.NoError(t, err)
	_, err = store.Get(ctx, number)
	require.Error(t, err)
	assert.EqualError(t, err, sql.ErrNoRows.Error())
}

func TestDeleteUncorrectStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	ctx, cancelfunc := context.WithTimeout(context.Background(), TestExecTimeout)
	defer cancelfunc()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	number, err := store.Add(ctx, parcel)
	require.NoError(t, err)
	require.Greater(t, number, 0)

	parcel.Number = number
	err = store.SetStatus(ctx, number, ParcelStatusSent)
	require.NoError(t, err)

	err = store.Delete(ctx, number)
	require.Error(t, err)
	assert.EqualError(t, err, ErrNoRowsDeleted.Error())
	p, err := store.Get(ctx, number)
	require.NoError(t, err)
	require.NotEmpty(t, p)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	ctx, cancelfunc := context.WithTimeout(context.Background(), TestExecTimeout)
	defer cancelfunc()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	number, err := store.Add(ctx, parcel)
	require.NoError(t, err)
	require.Greater(t, number, 0)

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(ctx, number, newAddress)
	require.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	p, err := store.Get(ctx, number)
	require.NoError(t, err)
	require.Equal(t, newAddress, p.Address)
}

func TestSetAddressUncorrectStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	ctx, cancelfunc := context.WithTimeout(context.Background(), TestExecTimeout)
	defer cancelfunc()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	number, err := store.Add(ctx, parcel)
	require.NoError(t, err)
	require.Greater(t, number, 0)

	parcel.Number = number
	err = store.SetStatus(ctx, number, ParcelStatusSent)
	require.NoError(t, err)

	newAddress := "new test address"
	err = store.SetAddress(ctx, number, newAddress)
	require.Error(t, err)
	assert.EqualError(t, err, ErrNoRowsUpdated.Error())

	p, err := store.Get(ctx, number)
	require.NoError(t, err)
	require.Equal(t, parcel.Address, p.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	ctx, cancelfunc := context.WithTimeout(context.Background(), TestExecTimeout)
	defer cancelfunc()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	number, err := store.Add(ctx, parcel)
	require.NoError(t, err)
	require.Greater(t, number, 0)

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	err = store.SetStatus(ctx, number, ParcelStatusSent)
	require.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	p, err := store.Get(ctx, number)
	require.NoError(t, err)
	require.Equal(t, ParcelStatusSent, p.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	ctx, cancelfunc := context.WithTimeout(context.Background(), TestExecTimeout)
	defer cancelfunc()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(ctx, parcels[i]) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		require.NoError(t, err)
		require.Greater(t, id, 0)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id
	}

	// get by client
	storedParcels, err := store.GetByClient(ctx, client) // получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))

	// check
	assert.ElementsMatch(t, parcels, storedParcels)
}
