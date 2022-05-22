package fancyindex

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (f *FileServer) directoryListing(files []os.FileInfo, canGoUp bool, root, urlPath string) browseTemplateContext {
	var dirCount, fileCount int
	var fileInfos []fileInfo
	for _, file := range files {
		name := file.Name()
		isDir := file.IsDir() || isSymlinkTargetDir(file, root, urlPath)

		// add the slash after the escape of path to avoid escaping the slash as well
		if isDir {
			name += "/"
			dirCount++
		} else {
			fileCount++
		}

		size := file.Size()
		fileIsSymlink := isSymlink(file)
		if fileIsSymlink {
			_path := sanitizedPathJoin(root, path.Join(urlPath, file.Name()))
			fileInfo, err := os.Lstat(_path)
			if err == nil {
				size = fileInfo.Size()
			}
		}

		u := url.URL{Path: "./" + name}

		fileInfos = append(fileInfos, fileInfo{
			IsDir:     isDir,
			IsSymlink: fileIsSymlink,
			Name:      name,
			Size:      size,
			URL:       fmt.Sprintf("%s?timestamp=%d", u.String(), time.Now().Unix()),
			ModTime:   file.ModTime().UTC(),
			Mode:      file.Mode(),
		})
	}
	name, _ := url.PathUnescape(urlPath)
	return browseTemplateContext{
		Auth:      f.auth,
		Timestamp: time.Now().Unix(),
		Name:      path.Base(name),
		Path:      urlPath,
		CanGoUp:   canGoUp,
		Items:     fileInfos,
		NumDirs:   dirCount,
		NumFiles:  fileCount,
	}
}

// browseTemplateContext provides the template context for directory listings.
type browseTemplateContext struct {
	// The authorizer for the current request.
	Auth bool `json:"Auth"`

	// The timestamp of the current time.
	Timestamp int64 `json:"Timestamp"`

	// The name of the directory (the last element of the path).
	Name string `json:"Name"`

	// The full path of the request.
	Path string `json:"Path"`

	CanGoUp bool `json:"CanGoUp"`

	// The items (files and folders) in the path.
	Items []fileInfo `json:"Items,omitempty"`

	// If ≠0 then Items starting from that many elements.
	Offset int `json:"Offset,omitempty"`

	// If ≠0 then Items have been limited to that many elements.
	Limit int `json:"Limit,omitempty"`

	// The number of directories in the listing.
	NumDirs int `json:"NumDirs"`

	// The number of files (items that aren't directories) in the listing.
	NumFiles int `json:"NumFiles"`

	// Sort column used
	Sort string `json:"Sort,omitempty"`

	// Sorting order
	Order string `json:"Order,omitempty"`
}

// Breadcrumbs returns l.Path where every element maps
// the link to the text to display.
func (l browseTemplateContext) Breadcrumbs() []crumb {
	if len(l.Path) == 0 {
		return []crumb{}
	}

	// skip trailing slash
	lpath := l.Path
	if lpath[len(lpath)-1] == '/' {
		lpath = lpath[:len(lpath)-1]
	}
	parts := strings.Split(lpath, "/")
	result := make([]crumb, len(parts))
	for i, p := range parts {
		if i == 0 && p == "" {
			p = "/"
		}
		p, _ = url.PathUnescape(p)
		lnk := strings.Repeat("../", len(parts)-i-1)
		lnk += fmt.Sprintf("?timestamp=%d", time.Now().Unix())
		result[i] = crumb{Link: lnk, Text: p}
	}
	return result
}

func (l *browseTemplateContext) applySortAndLimit(sortParam, orderParam, limitParam string, offsetParam string) {
	l.Sort = sortParam
	l.Order = orderParam

	if l.Order == "desc" {
		switch l.Sort {
		case sortByName:
			sort.Sort(sort.Reverse(byName(*l)))
		case sortByNameDirFirst:
			sort.Sort(sort.Reverse(byNameDirFirst(*l)))
		case sortBySize:
			sort.Sort(sort.Reverse(bySize(*l)))
		case sortByTime:
			sort.Sort(sort.Reverse(byTime(*l)))
		}
	} else {
		switch l.Sort {
		case sortByName:
			sort.Sort(byName(*l))
		case sortByNameDirFirst:
			sort.Sort(byNameDirFirst(*l))
		case sortBySize:
			sort.Sort(bySize(*l))
		case sortByTime:
			sort.Sort(byTime(*l))
		}
	}

	if offsetParam != "" {
		offset, _ := strconv.Atoi(offsetParam)
		if offset > 0 && offset <= len(l.Items) {
			l.Items = l.Items[offset:]
			l.Offset = offset
		}
	}

	if limitParam != "" {
		limit, _ := strconv.Atoi(limitParam)

		if limit > 0 && limit <= len(l.Items) {
			l.Items = l.Items[:limit]
			l.Limit = limit
		}
	}
}

// crumb represents part of a breadcrumb menu,
// pairing a link with the text to display.
type crumb struct {
	Link, Text string
}

// fileInfo contains serializable information
// about a file or directory.
type fileInfo struct {
	Name      string      `json:"name"`
	Size      int64       `json:"size"`
	URL       string      `json:"url"`
	ModTime   time.Time   `json:"mod_time"`
	Mode      os.FileMode `json:"mode"`
	IsDir     bool        `json:"is_dir"`
	IsSymlink bool        `json:"is_symlink"`
}

// HumanSize returns the size of the file as a
// human-readable string in IEC format (i.e.
// power of 2 or base 1024).
func (fi fileInfo) HumanSize() string {
	return humanize.IBytes(uint64(fi.Size))
}

// HumanModTime returns the modified time of the file
// as a human-readable string given by format.
func (fi fileInfo) HumanModTime(format string) string {
	return fi.ModTime.Format(format)
}

type (
	byName         browseTemplateContext
	byNameDirFirst browseTemplateContext
	bySize         browseTemplateContext
	byTime         browseTemplateContext
)

func (l byName) Len() int      { return len(l.Items) }
func (l byName) Swap(i, j int) { l.Items[i], l.Items[j] = l.Items[j], l.Items[i] }

func (l byName) Less(i, j int) bool {
	return strings.ToLower(l.Items[i].Name) < strings.ToLower(l.Items[j].Name)
}

func (l byNameDirFirst) Len() int      { return len(l.Items) }
func (l byNameDirFirst) Swap(i, j int) { l.Items[i], l.Items[j] = l.Items[j], l.Items[i] }

func (l byNameDirFirst) Less(i, j int) bool {
	// sort by name if both are dir or file
	if l.Items[i].IsDir == l.Items[j].IsDir {
		return strings.ToLower(l.Items[i].Name) < strings.ToLower(l.Items[j].Name)
	}
	// sort dir ahead of file
	return l.Items[i].IsDir
}

func (l bySize) Len() int      { return len(l.Items) }
func (l bySize) Swap(i, j int) { l.Items[i], l.Items[j] = l.Items[j], l.Items[i] }

func (l bySize) Less(i, j int) bool {
	const directoryOffset = -1 << 31 // = -math.MinInt32

	iSize, jSize := l.Items[i].Size, l.Items[j].Size

	// directory sizes depend on the file system; to
	// provide a consistent experience, put them up front
	// and sort them by name
	if l.Items[i].IsDir {
		iSize = directoryOffset
	}
	if l.Items[j].IsDir {
		jSize = directoryOffset
	}
	if l.Items[i].IsDir && l.Items[j].IsDir {
		return strings.ToLower(l.Items[i].Name) < strings.ToLower(l.Items[j].Name)
	}

	return iSize < jSize
}

func (l byTime) Len() int           { return len(l.Items) }
func (l byTime) Swap(i, j int)      { l.Items[i], l.Items[j] = l.Items[j], l.Items[i] }
func (l byTime) Less(i, j int) bool { return l.Items[i].ModTime.Before(l.Items[j].ModTime) }

const (
	sortByName         = "name"
	sortByNameDirFirst = "namedirfirst"
	sortBySize         = "size"
	sortByTime         = "time"
)

// isSymlink return true if f is a symbolic link
func isSymlink(f os.FileInfo) bool {
	return f.Mode()&os.ModeSymlink == os.ModeSymlink
}

// isSymlinkTargetDir returns true if f's symbolic link target
// is a directory.
func isSymlinkTargetDir(f os.FileInfo, root, urlPath string) bool {
	if !isSymlink(f) {
		return false
	}
	target := sanitizedPathJoin(root, path.Join(urlPath, f.Name()))
	targetInfo, err := os.Stat(target)
	if err != nil {
		return false
	}
	return targetInfo.IsDir()
}

const separator = string(filepath.Separator)

func sanitizedPathJoin(root, reqPath string) string {
	if root == "" {
		root = "."
	}

	_path := filepath.Join(root, filepath.Clean("/"+reqPath))

	// filepath.Join also cleans the path, and cleaning strips
	// the trailing slash, so we need to re-add it afterwards.
	// if the length is 1, then it's a path to the root,
	// and that should return ".", so we don't append the separator.
	if strings.HasSuffix(reqPath, "/") && len(reqPath) > 1 {
		_path += separator
	}

	return _path
}

func (f *FileServer) calculateAbsolutePath(root string) string {
	absolutePath := path.Join("/", root)
	if len(absolutePath) > 0 && absolutePath[len(absolutePath)-1] == '/' {
		absolutePath = absolutePath[:len(absolutePath)-1]
	}
	return absolutePath
}
