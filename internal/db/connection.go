package db


import ("github.com/jackc/pgx/v5/pgxpool"
   "context"
       "os"
)
func NewPool()(*pgxpool.Pool,error) {
    connstring:=os.Getenv("DATABASE_URL")
	pool,err:=pgxpool.New(context.Background(),connstring)
if err!=nil{
	return nil,err
}

return pool,nil
}
