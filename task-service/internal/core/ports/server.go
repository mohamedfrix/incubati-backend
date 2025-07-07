package ports



type Server interface {
	Start(any) error
}