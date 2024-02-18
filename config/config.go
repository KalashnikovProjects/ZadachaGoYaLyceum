package config

// Operation - ORM для базы данных sqlite и общее представление операции (x + y), мы храним состояние для каждого финального, начального и промежуточного вычисления
// финальное вычисление - 2 готовых числа, промежуточное
// Всех кроме финального имеют не LeftData или RightData и LeftIsReady, RightIsReady = 0, они ждут пока их изменят более
// низкоуровневые (финальные вычисления) и так рекурсивно оно решается.
type Operation struct {
	Id               int `gorm:"unique; primaryKey; autoIncrement"`
	Znak             string
	LeftIsReady      int
	LeftData         float64
	RightIsReady     int
	RightData        float64
	FatherId         int
	SonSide          int // 0 или 1 (левая или правая векта) к которой крепится выражение
	FinalOperationId int // id из FinalOperation
}

type FinalOperation struct {
	Id        int `gorm:"unique; primaryKey; autoIncrement"`
	NeedToDo  string
	Status    string  // error / process / done
	Result    float64 // если статус done, иначе -1
	StartTime int
	EndTime   int
}

type Agent struct {
	Id         int    `gorm:"unique; primaryKey; autoIncrement"`
	Status     string // error / process / waiting    error только если агент не смог инициализироваться
	StatusText string // сообщение о ошибке или выполняемое выражение
	PingTime   int
}

type OperationTime struct {
	Id   int `gorm:"unique; primaryKey"` // 1 = +   2 = -   3 = *   4 = /
	Time int // В секундах
}

//  Упаковать всё в docker  // готово
//  Баги какие то   // пофиксил
//  Красивые схемки (готово) и прочитай меня Readme.md !!!!!!!!!
//  Грузим на GitHub (ПУБЛИЧНЫЙ РЕПОЗИТОРИЙ!!!!)
//  Передобавление в очередь при ошибке в agent (не успеваю)
