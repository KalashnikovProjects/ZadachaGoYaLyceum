package entities

// Operation - ORM для базы данных sqlite и общее представление операции (x + y), мы храним состояние для каждого финального, начального и промежуточного вычисления
// финальное вычисление - 2 готовых числа, промежуточное
// Всех кроме финального имеют не LeftData или RightData и LeftIsReady, RightIsReady = 0, они ждут пока их изменят более
// низкоуровневые (финальные вычисления) и так рекурсивно оно решается.
type Operation struct {
	Id           int     `json:"id"`
	Znak         string  `json:"znak"`
	LeftIsReady  int     `json:"-"`
	LeftData     float64 `json:"-"`
	RightIsReady int     `json:"-"`
	RightData    float64 `json:"-"`
	FatherId     int     `json:"-"`
	SonSide      int     `json:"-"`             // 0 или 1 (левая или правая векта) к которой крепится выражение
	ExpressionId int     `json:"expression_id"` // id из Expression
}

type Expression struct {
	Id        int     `json:"id"`
	NeedToDo  string  `json:"need_to_do"`
	Status    string  `json:"status"` // error / process / done
	Result    float64 `json:"result"` // если статус done, иначе -1
	StartTime int     `json:"start_time"`
	EndTime   int     `json:"end_time"`
	UserId    int     `json:"user_id"`
}

type Agent struct {
	Id         int    `json:"id"`
	Status     string `json:"status"`      // error / process / waiting    my_errors только если агент не смог инициализироваться
	StatusText string `json:"status_text"` // сообщение о ошибке или выполняемое выражение
	PingTime   int    `json:"ping_time"`
}

type User struct {
	Id               int    `json:"id"`
	Name             string `json:"name"`
	Password         string `json:"password"`
	PasswordHash     string `json:"-"`
	OperationsTimeId int    `json:"-"`
}

type OperationsTime struct {
	Id             int
	Plus           int `json:"plus"` // В секундах
	Minus          int `json:"minus"`
	Division       int `json:"division"`
	Multiplication int `json:"multiplication"`
}

// IdSoup Хз зачем, но пусть будет
type IdSoup struct {
	OperationId      int
	ExpressionId     int
	UserId           int
	OperationsTimeId int
}
