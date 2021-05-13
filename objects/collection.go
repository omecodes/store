package objects

import (
	"context"
	pb "github.com/omecodes/store/gen/go/proto"
	"sync"
)

type CollectionDB interface {

	// Save saves object content as JSON in database
	Save(ctx context.Context, object *pb.Object, index ...*pb.TextIndex) error

	// Patch applies patch to object with id equals patch.ID()
	Patch(ctx context.Context, patch *pb.Patch) error

	// Delete removes all content associated with objectID
	Delete(ctx context.Context, objectID string) error

	// List returns a list of at most 'opts.Count' objects
	List(ctx context.Context, opts ListOptions) (*Cursor, error)

	// Get gets the object associated with objectID
	Get(ctx context.Context, objectID string, opts GetObjectOptions) (*pb.Object, error)

	// Info gets header of the object associated with objectID
	Info(ctx context.Context, objectID string) (*pb.Header, error)

	Search(ctx context.Context, query *pb.SearchQuery) (*Cursor, error)

	// Clear removes all objects store
	Clear() error
}

type collectionContainer struct {
	sync.RWMutex
	container map[string]CollectionDB
}

func (c *collectionContainer) Get(name string) (CollectionDB, bool) {
	c.RLock()
	defer c.RUnlock()
	collection, found := c.container[name]
	return collection, found
}

func (c *collectionContainer) Save(name string, collection CollectionDB) {
	c.Lock()
	defer c.Unlock()
	c.container[name] = collection
}

func (c *collectionContainer) Delete(name string) {
	c.Lock()
	defer c.Unlock()
	delete(c.container, name)
}
