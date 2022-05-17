package outputs

type Client interface {
	Close() error

	Connect() error

	Publish() error
}
