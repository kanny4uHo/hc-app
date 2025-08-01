package entity

type AddUserArgs struct {
	Login    string
	Password string
	Meta     UserMeta
}

type UserShort struct {
	ID    int64
	Login string
}

type User struct {
	UserShort
	PasswordHash string
	Meta         UserMeta
}

type UserMeta struct {
	Name  UserName
	Email string
}

type UserName struct {
	First string
	Last  string
}

func (user User) IsEmpty() bool {
	return user.ID == 0
}
