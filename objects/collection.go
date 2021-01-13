package objects

import (
	"context"
	"github.com/omecodes/store/pb"
	"sync"
)

type Collection interface {

	// Save saves object content as JSON in database
	Save(ctx context.Context, object *pb.Object, index ...*pb.TextIndex) error

	// Update applies patch to object with id equals patch.ID()
	Patch(ctx context.Context, patch *pb.Patch) error

	// Delete removes all content associated with objectID
	Delete(ctx context.Context, objectID string) error

	// List returns a list of at most 'opts.Count' objects
	List(ctx context.Context, opts pb.ListOptions) (*pb.Cursor, error)

	// Get gets the object associated with objectID
	Get(ctx context.Context, objectID string, opts pb.GetOptions) (*pb.Object, error)

	// Info gets header of the object associated with objectID
	Info(ctx context.Context, objectID string) (*pb.Header, error)

	// Search
	Search(ctx context.Context, query *pb.SearchQuery) (*pb.Cursor, error)

	// Clear removes all objects store
	Clear() error
}

type collectionContainer struct {
	sync.RWMutex
	container map[string]Collection
}

func (c *collectionContainer) Get(name string) (Collection, bool) {
	c.RLock()
	defer c.RUnlock()
	collection, found := c.container[name]
	return collection, found
}

func (c *collectionContainer) Save(name string, collection Collection) {
	c.Lock()
	defer c.Unlock()
	c.container[name] = collection
}

func (c *collectionContainer) Delete(name string) {
	c.Lock()
	defer c.Unlock()
	delete(c.container, name)
}
