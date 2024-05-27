package discovery

import (
	"net/url"

	"github.com/dogsays/mo/ut2"

	"google.golang.org/grpc/resolver"
)

type FileStaticDiscovery struct {
	*FilePortProvider
	list *ut2.Array[*fileResolver]
}

func NewFileStaticDiscovery(filename string) *FileStaticDiscovery {
	fd := &FileStaticDiscovery{}
	fd.list = &ut2.Array[*fileResolver]{}

	fd.FilePortProvider = NewFilePortProvider(filename)

	return fd
}

func (mb *FileStaticDiscovery) Scheme() string {
	return "file"
}

func (f *FileStaticDiscovery) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	r := newfileResolver(target.URL, cc, f)
	r.ResolveNow(resolver.ResolveNowOptions{})
	f.list.PushBack(r)
	return r, nil
}

type fileResolver struct {
	URL  url.URL
	cc   resolver.ClientConn
	file *FileStaticDiscovery
}

func newfileResolver(u url.URL, cc resolver.ClientConn, f *FileStaticDiscovery) *fileResolver {
	return &fileResolver{URL: u, cc: cc, file: f}
}

func (r *fileResolver) Close() {
	r.file.list.RemoveByValue(r)
}

func (r *fileResolver) ResolveNow(opt resolver.ResolveNowOptions) {
	r.cc.UpdateState(resolver.State{Addresses: r.build()})
}

func (r *fileResolver) build() (ret []resolver.Address) {

	// scheme://host/path
	serviceName := r.URL.Host

	arr, _ := r.file.GetAddr(serviceName)
	if len(arr) > 0 {
		for _, v := range arr {
			ret = append(ret, resolver.Address{Addr: v})
		}
	} else {
		// for debug
		ret = append(ret, resolver.Address{Addr: r.URL.String()})
	}

	return
}
