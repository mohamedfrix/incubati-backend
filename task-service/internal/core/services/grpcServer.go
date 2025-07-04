package services


type GRPCServer struct {
	Port     string
}

func NewGRPCServer(port string) *GRPCServer {
	return &GRPCServer{
		Port: port,
	}
}


func (s *GRPCServer) Start(params any) error {


	return nil
}