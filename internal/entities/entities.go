package entities

// TODO:
//    2. Middleware для user_server, проверяет session storage, если плохо - редирект на /login
//    3. Добавить user id во все запросы к expressions, operations, доделать operations_time
//  2.1  Убрать rabbitMQ, центральная горутина Агент принимает по GRPC заказы и каналами распределяет по агентикам
//  3.1  Добавить микро тестов
//  3.2  Добавить 2-3 гига тестов всего и вся

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
	FinalOperationId int // id из Expression
}

type Expression struct {
	Id        int `gorm:"unique; primaryKey; autoIncrement"`
	NeedToDo  string
	Status    string  // my_errors / process / done
	Result    float64 // если статус done, иначе -1
	StartTime int
	EndTime   int
	UserId    int
}

type Agent struct {
	Id         int    `gorm:"unique; primaryKey; autoIncrement"`
	Status     string // my_errors / process / waiting    my_errors только если агент не смог инициализироваться
	StatusText string // сообщение о ошибке или выполняемое выражение
	PingTime   int
}

type User struct {
	Id               int `gorm:"unique; primaryKey; autoIncrement"`
	PasswordHash     string
	OperationsTimeId int
}

type OperationsTime struct {
	Id             int `gorm:"unique; primaryKey"`
	Plus           int `json:"plus"` // В секундах
	Minus          int `json:"minus"`
	Division       int `json:"division"`
	Multiplication int `json:"multiplication"`
}
