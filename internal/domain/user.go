package domain

// User 领域对象，是DDD中的聚合根
// BO(business object)
type User struct {
	Email    string
	Password string
}
