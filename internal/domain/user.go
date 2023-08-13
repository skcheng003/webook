package domain

// User 领域对象，是DDD中的聚合根
// BO(business object)
type User struct {
	Id       int64
	Email    string
	Password string
}
