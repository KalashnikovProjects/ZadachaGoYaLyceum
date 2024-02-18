package db_connect

import (
	"Zadacha/config"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

// Пакет для работы с базой данных, использует GORM

type DBConnection struct {
	db *gorm.DB
}

func OpenDb(connectionString string) (DBConnection, error) {
	// Подключение к базе данных PostgreSQL
	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
		return DBConnection{}, err
	}

	// Автомиграция модели config.Operation
	err = db.AutoMigrate(&config.Operation{}, &config.Agent{}, &config.FinalOperation{}, &config.OperationTime{})
	if err != nil {
		fmt.Println("Ошибка при миграции таблиц:", err)
		return DBConnection{}, err
	}
	res := DBConnection{db: db}
	res.InitOperationTime()
	return res, nil
}

func (conn *DBConnection) Get(id int) (config.Operation, error) {
	var data config.Operation
	result := conn.db.First(&data, id)
	if result.Error != nil {
		return config.Operation{}, result.Error
	}
	return data, nil
}

func (conn *DBConnection) Add(value *config.Operation) (int, error) {
	result := conn.db.Create(value)
	if result.Error != nil {
		return 0, result.Error
	}
	return value.Id, nil
}

func (conn *DBConnection) GetAll() []config.Operation {
	var res []config.Operation
	conn.db.Find(&res)
	return res
}

func (conn *DBConnection) UpdateLeft(id int, leftData float64) error {
	err := conn.db.Transaction(func(tx *gorm.DB) error {
		query := `
    UPDATE operations
    SET left_data = $1,
        left_is_ready = $2
    WHERE id = $3
`
		result := conn.db.Exec(query, leftData, 1, id)
		return result.Error
	})
	return err
	return err
}

func (conn *DBConnection) UpdateRight(id int, rightData float64) error {
	err := conn.db.Transaction(func(tx *gorm.DB) error {
		query := `
    UPDATE operations
    SET right_data = $1,
        right_is_ready = $2
    WHERE id = $3
`
		result := conn.db.Exec(query, rightData, 1, id)
		return result.Error
	})
	return err
}

func (conn *DBConnection) UpdateFather(id, FatherID, side int) error {
	var value config.Operation
	result := conn.db.First(&value, id)
	if result.Error != nil {
		return result.Error
	}

	value.FatherId = FatherID
	value.SonSide = side
	result = conn.db.Save(&value)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (conn *DBConnection) Delete(id int) error {
	var value config.Operation
	result := conn.db.First(&value, id)
	if result.Error != nil {
		return result.Error
	}

	result = conn.db.Delete(&value)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (conn *DBConnection) IsReadyToExecute(id int) (bool, error) {
	var value config.Operation
	result := conn.db.First(&value, id)
	if result.Error != nil {
		return false, result.Error
	}
	if value.LeftIsReady == 1 && value.RightIsReady == 1 {
		return true, nil
	}
	return false, nil
}

func (conn *DBConnection) AgentPing(id int, status, statusText string) int {
	var value config.Agent

	if status == "create" {
		data := config.Agent{Status: status, PingTime: int(time.Now().Unix())}
		conn.db.Create(&data)
		return data.Id
	}
	result := conn.db.First(&value, id)
	if result.Error != nil {
		data := config.Agent{Status: status, PingTime: int(time.Now().Unix())}
		conn.db.Create(&data)
		return data.Id
	}
	value.PingTime = int(time.Now().Unix())
	value.Status = status
	value.StatusText = statusText
	result = conn.db.Save(&value)
	return value.Id
}

func (conn *DBConnection) DeleteAgent(id int) error {
	var value config.Agent
	result := conn.db.First(&value, id)
	if result.Error != nil {
		return result.Error
	}

	result = conn.db.Delete(&value)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (conn *DBConnection) GetAllAgents() []config.Agent {
	var res []config.Agent
	conn.db.Find(&res)
	return res
}

func (conn *DBConnection) CreateFinalOperation(operation config.FinalOperation) (int, error) {
	result := conn.db.Create(&operation)
	if result.Error != nil {
		return 0, result.Error
	}
	return operation.Id, nil
}

func (conn *DBConnection) GetFinalOperationByID(id int) (config.FinalOperation, error) {
	var data config.FinalOperation
	result := conn.db.First(&data, id)
	if result.Error != nil {
		return config.FinalOperation{}, result.Error
	}
	return data, nil
}

func (conn *DBConnection) GetAllFinalOperations() []config.FinalOperation {
	var res []config.FinalOperation
	conn.db.Find(&res)
	return res
}

func (conn *DBConnection) UpdateFinalOperation(id int, newResult float64, status string) error {
	var value config.FinalOperation
	result := conn.db.First(&value, id)
	if result.Error != nil {
		return result.Error
	}

	value.Status = status
	value.EndTime = int(time.Now().Unix())
	if status == "done" {
		value.Result = newResult
	}
	result = conn.db.Save(&value)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (conn *DBConnection) OhNoFinalOperationError(id int) {
	var value config.FinalOperation
	result := conn.db.First(&value, id)
	if result.Error != nil {
		return
	}

	value.Status = "error"
	value.EndTime = int(time.Now().Unix())
	conn.db.Where("final_operation_id = ?", id).Delete(&config.Operation{})
	result = conn.db.Save(&value)
	if result.Error != nil {
		return
	}
	return
}

func (conn *DBConnection) Close() error {
	db, err := conn.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

func (conn *DBConnection) InitOperationTime() {
	for id := 1; id < 5; id++ {
		var a config.OperationTime
		result := conn.db.First(&a, id)
		if result.Error != nil {
			conn.db.Create(&config.OperationTime{Id: id, Time: 10})
		}
	}
}

func (conn *DBConnection) GetAllOperationTime() []config.OperationTime {
	var res []config.OperationTime
	conn.db.Find(&res)
	return res
}

func (conn *DBConnection) GetOperationTimeByID(id int) (config.OperationTime, error) {
	var data config.OperationTime
	result := conn.db.First(&data, id)
	if result.Error != nil {
		return config.OperationTime{}, result.Error
	}
	return data, nil
}

func (conn *DBConnection) UpdateOperationTime(id, time int) error {
	var value config.OperationTime
	result := conn.db.First(&value, id)
	if result.Error != nil {
		return result.Error
	}
	value.Time = time
	result = conn.db.Save(&value)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
