package tests

import (
	"bytes"
	"fmt"
	"github.com/lazychanger/filesystem"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"sync"
	"testing"
)

const (
	FuncOpen        = "Open"
	FuncOpenErr     = "OpenErr"
	FuncCreate      = "Create"
	FuncCreateErr   = "CreateErr"
	FuncMkdir       = "Mkdir"
	FuncMkdirErr    = "MkdirErr"
	FuncMkdirAll    = "MkdirAll"
	FuncRemove      = "Remove"
	FuncRemoveErr   = "RemoveErr"
	FuncRemoveAll   = "RemoveAll"
	FuncRename      = "Rename"
	FuncRenameErr   = "RenameErr"
	FuncSub         = "Sub"
	FuncSubErr      = "SubErr"
	FuncStat        = "Stat"
	FuncStatErr     = "StatErr"
	FuncExist       = "Exist"
	FuncExistFalse  = "ExistFalse"
	FuncIsFile      = "IsFile"
	FuncIsFileFalse = "IsFileFalse"
	FuncIsDir       = "IsDir"
	FuncIsDirFalse  = "IsDirFalse"

	//FuncReadDirFs   = "ReadDirFs"
	//FuncReadDir     = "ReadDir"
	//FuncReadFileFs  = "ReadFileFS"
	//FuncReadFile    = "ReadFile"
	//FuncOpenFileFs  = "OpenFileFs"
	//FuncOpenFile    = "OpenFile"
	//FuncWriteFileFs = "WriteFileFs"
)

type EventHandlerFunc func(path string) error

type EventRegisterFunc func(e *EventDispatcher)

type EventDispatcher struct {
	events map[string][]EventHandlerFunc

	sync.RWMutex
}

func (e *EventDispatcher) Listener(funcName string, eventFunc EventHandlerFunc) {
	e.Lock()
	defer e.Unlock()

	if _, ok := e.events[funcName]; !ok {
		e.events[funcName] = make([]EventHandlerFunc, 0)
	}

	e.events[funcName] = append(e.events[funcName], eventFunc)
}

func (e *EventDispatcher) call(funcName, path string, t *testing.T) {

	e.RLock()
	defer e.RUnlock()

	if handlers, ok := e.events[funcName]; ok && len(handlers) > 0 {
		for _, handler := range handlers {
			err := handler(path)
			assert.Equal(t, err == nil, true, fmt.Sprintf("funcName: %s, path: %s, err: %s", funcName, path, err))
		}
	}

}

func TestDriver(t *testing.T, dsn string, eventRegisters ...EventRegisterFunc) {
	var (
		vfs        filesystem.FileSystem
		err        error
		dispatcher = &EventDispatcher{
			events: make(map[string][]EventHandlerFunc),
		}
	)

	for _, eventRegister := range eventRegisters {
		eventRegister(dispatcher)
	}

	eventRegisters = nil

	vfs, err = filesystem.Open(dsn)
	if err != nil {
		t.Error(err)
		return
	}

	// default dir tree

	// |- test.txt
	// |- test_dir
	// |- - test1.txt
	// |- - test_dir2
	// |- - - test2.txt

	t.Run("init dir tree", func(t *testing.T) {
		_, err := vfs.Create("/test.txt")
		assert.NoError(t, err)
		dispatcher.call(FuncCreate, "/test.txt", t)

		assert.NoError(t, vfs.MkdirAll("/test_dir/test_dir2", 0755))
		dispatcher.call(FuncMkdir, "/test_dir/test_dir2", t)

		_, err = vfs.Create("/test_dir/test1.txt")
		assert.NoError(t, err)
		dispatcher.call(FuncCreate, "/test_dir/test1.txt", t)

		_, err = vfs.Create("/test_dir/test_dir2/test2.txt")
		assert.NoError(t, err)
		dispatcher.call(FuncCreate, "/test_dir/test_dir2/test2.txt", t)
	})

	t.Run("test Open", func(t *testing.T) {
		_, err := vfs.Open("/test.txt")
		assert.NoError(t, err)
		dispatcher.call(FuncOpen, "/test.txt", t)

		_, err = vfs.Open("/new_open.txt")
		assert.Error(t, err)
		dispatcher.call(FuncOpen, "/new_open.txt", t)

		t.Run("test Open error", func(t *testing.T) {
			_, err = vfs.Open("/noexist/test.txt")
			assert.Error(t, err)
			dispatcher.call(FuncOpenErr, "/noexist/test.txt", t)
		})
	})

	t.Run("test Create and Remove", func(t *testing.T) {
		_, err := vfs.Create("/test_create.txt")
		assert.NoError(t, err)
		dispatcher.call(FuncCreate, "/test_create.txt", t)

		_, err = vfs.Create("/test.txt")
		assert.NoError(t, err)
		dispatcher.call(FuncCreate, "/test.txt", t)

		t.Run("test Create error", func(t *testing.T) {
			_, err = vfs.Create("/noexist/test_create.txt")
			assert.Error(t, err)
			dispatcher.call(FuncCreateErr, "/test_create.txt", t)
		})

		assert.NoError(t, vfs.Remove("/test_create.txt"))
		dispatcher.call(FuncRemove, "/test_create.txt", t)

		assert.NoError(t, vfs.Mkdir("/remove_dir", 0755))
		dispatcher.call(FuncMkdir, "/remove_dir", t)

		assert.NoError(t, vfs.Remove("/remove_dir/"))
		dispatcher.call(FuncRemove, "/remove_dir/", t)

		_, err = vfs.Stat("/remove_dir")
		assert.Error(t, err)

		assert.NoError(t, vfs.Mkdir("/remove_dir2", 0755))
		dispatcher.call(FuncMkdir, "/remove_dir2", t)

		assert.NoError(t, vfs.Remove("/remove_dir2"))
		dispatcher.call(FuncRemove, "/remove_dir2", t)

		_, err = vfs.Stat("/remove_dir")
		assert.Error(t, err)

		t.Run("test Remove Error", func(t *testing.T) {
			assert.Error(t, vfs.Remove("/noexist/test_create.txt"))
			dispatcher.call(FuncRemoveErr, "/noexist/test_create.txt", t)

			assert.Error(t, vfs.Remove("/noexist/"))
			dispatcher.call(FuncRemoveErr, "/noexist/", t)

			assert.Error(t, vfs.Remove("/noexist"))
			dispatcher.call(FuncRemoveErr, "/noexist", t)

		})
	})

	t.Run("test Mkdir", func(t *testing.T) {
		assert.NoError(t, vfs.Mkdir("/test_dir_mkdir", 0777))
		// if dir create succeed, then dir should be existed
		_, err = vfs.Create("/test_dir_mkdir/test.txt")
		assert.NoError(t, err)

		dispatcher.call(FuncMkdir, "/test_dir_mkdir", t)

		assert.Error(t, vfs.Mkdir("/test_dir", 0755))
		dispatcher.call(FuncMkdir, ".test_dir", t)

		t.Run("test Mkdir error", func(t *testing.T) {
			assert.Error(t, vfs.Mkdir("/test1/noexist", 0777))
			dispatcher.call(FuncMkdirErr, "/test1/noexist", t)
		})

		assert.NoError(t, vfs.RemoveAll("/test_dir_mkdir"))
	})

	t.Run("test MkdirAll and RemoveAll", func(t *testing.T) {
		assert.NoError(t, vfs.MkdirAll("/test_dir_mkdirall/test/test3", 0777))
		dispatcher.call(FuncMkdirAll, "/test_dir_mkdirall/test/test3", t)

		assert.NoError(t, vfs.RemoveAll("/test_dir_mkdirall"))
		dispatcher.call(FuncRemoveAll, "/test_dir_mkdirall", t)

		assert.NoError(t, vfs.RemoveAll("/noexist"))
		dispatcher.call(FuncRemoveAll, "/noexist", t)

		assert.NoError(t, vfs.RemoveAll("/noexist/"))
		dispatcher.call(FuncRemoveAll, "/noexist/", t)

		assert.NoError(t, vfs.RemoveAll("/noexist/noexist2"))
		dispatcher.call(FuncRemoveAll, "/noexist/noexist2", t)

		assert.NoError(t, vfs.RemoveAll("/noexist/noexist2/"))
		dispatcher.call(FuncRemoveAll, "/noexist/noexist2/", t)
	})

	t.Run("test State", func(t *testing.T) {
		fi, err := vfs.Stat("/test.txt")
		assert.NoError(t, err)
		if fi != nil {
			assert.Equal(t, fi.Name(), "test.txt")
			assert.Equal(t, fi.IsDir(), false)
			dispatcher.call(FuncStat, "/test.txt", t)
		}

		fi, err = vfs.Stat("/test_dir")
		assert.NoError(t, err)
		if fi != nil {
			assert.Equal(t, fi.Name(), "test_dir")
			assert.Equal(t, fi.IsDir(), true)
			dispatcher.call(FuncStat, "/test_dir", t)
		}

		t.Run("test State error", func(t *testing.T) {
			_, err := vfs.Stat("/noexist/test.txt")
			assert.Error(t, err)
			dispatcher.call(FuncStatErr, "/noexist/test.txt", t)
		})
	})

	t.Run("test Exists", func(t *testing.T) {
		assert.True(t, vfs.Exists("/test.txt"))
		dispatcher.call(FuncExist, "/test.txt", t)

		assert.True(t, vfs.Exists("/test_dir/test_dir2"))
		dispatcher.call(FuncExist, "/test_dir/test_dir2", t)

		assert.False(t, vfs.Exists("/test_dir/noexist"))
		dispatcher.call(FuncExistFalse, "/test_dir/noexist", t)

		assert.False(t, vfs.Exists("/noexist"))
		dispatcher.call(FuncExistFalse, "/noexist", t)

		assert.False(t, vfs.Exists("/noexist/noexist2"))
		dispatcher.call(FuncExistFalse, "/noexist/noexist", t)
	})

	t.Run("test IsFile", func(t *testing.T) {
		assert.True(t, vfs.IsFile("/test.txt"))
		dispatcher.call(FuncIsFile, "/test.txt", t)

		assert.True(t, vfs.IsFile("/test_dir/test1.txt"))
		dispatcher.call(FuncIsFile, "/test_dir/test1.txt", t)

		assert.True(t, vfs.IsFile("/test_dir/test_dir2/test2.txt"))
		dispatcher.call(FuncIsFile, "/test_dir/test_dir2/test2.txt", t)

		assert.False(t, vfs.IsFile("/test_dir"))
		dispatcher.call(FuncIsFileFalse, "/test_dir", t)

		assert.False(t, vfs.IsFile("/noexist"))
		dispatcher.call(FuncIsFileFalse, "/noexist", t)

		assert.False(t, vfs.IsFile("/noexist/noexist.txt"))
		dispatcher.call(FuncIsFileFalse, "/noexist/noexist.txt", t)
	})

	t.Run("test IsDir", func(t *testing.T) {
		assert.True(t, vfs.IsDir("/test_dir"))
		dispatcher.call(FuncIsDir, "/test_dir", t)

		assert.True(t, vfs.IsDir("/test_dir/test_dir2"))
		dispatcher.call(FuncIsDir, "/test_dir/test_dir2", t)

		assert.False(t, vfs.IsDir("/noexist"))
		dispatcher.call(FuncIsDirFalse, "/noexist", t)

		assert.False(t, vfs.IsDir("/noexist/noexist2"))
		dispatcher.call(FuncIsDirFalse, "/noexist/noexist2", t)
	})

	t.Run("test Rename", func(t *testing.T) {

		t.Run("test Rename file", func(t *testing.T) {
			var (
				err error
				fi  fs.FileInfo
			)
			_, err = vfs.Create("/test_dir/rename.txt")
			assert.NoError(t, err)
			dispatcher.call(FuncCreate, "/test_dir/rename.txt", t)

			_, err = vfs.Create("/test_dir/rename_cover.txt")
			assert.NoError(t, err)
			dispatcher.call(FuncCreate, "/test_dir/rename_cover.txt", t)

			assert.NoError(t, vfs.Rename("/test_dir/rename.txt", "/test_dir/rename_cover.txt"))
			dispatcher.call(FuncRenameErr, "/test_dir/rename.txt:/test_dir/rename_cover.txt", t)

			assert.NoError(t, vfs.Rename("/test_dir/rename_cover.txt", "/test_dir/test_dir2/rename2.txt"))
			dispatcher.call(FuncRename, "/test_dir/rename_cover.txt:/test_dir/test_dir2/rename2.txt", t)

			fi, err = vfs.Stat("/test_dir/test_dir2/rename2.txt")
			assert.NoError(t, err)
			if fi != nil {
				assert.Equal(t, fi.Name(), "rename2.txt")
				assert.False(t, fi.IsDir())
			}

			_, err = vfs.Stat("/test_dir/rename.txt")
			assert.Error(t, err)

			assert.NoError(t, vfs.RemoveAll("/test_dir/test_dir2/rename2.txt"))
		})

		t.Run("test Rename dir", func(t *testing.T) {
			var (
				err error
				fi  fs.FileInfo
			)

			assert.Error(t, vfs.Rename("/test_dir/noexist/", "/test_dir/rename_dir5"))
			_, err = vfs.Stat("/test_dir/rename_dir5")
			assert.Error(t, err)

			assert.NoError(t, vfs.Mkdir("/test_dir/rename_dir", 0755))
			dispatcher.call(FuncMkdir, "/test_dir/rename_dir", t)

			assert.Error(t, vfs.Rename("/test_dir/rename_dir", "/test_dir/test_dir2"))
			dispatcher.call(FuncMkdirErr, "/test_dir/rename_dir:/test_dir/test_dir2", t)

			assert.NoError(t, vfs.Rename("/test_dir/rename_dir", "/test_dir/test_dir2/rename_dir2"))
			dispatcher.call(FuncRename, "/test_dir/rename_dir:/test_dir/test_dir2/rename_dir2", t)

			fi, err = vfs.Stat("/test_dir/test_dir2/rename_dir2")
			assert.NoError(t, err)
			if fi != nil {
				assert.Equal(t, fi.Name(), "rename_dir2")
				assert.True(t, fi.IsDir())
			}

			_, err = vfs.Stat("/test_dir/rename_dir")
			assert.Error(t, err)

			assert.NoError(t, vfs.Rename("/test_dir/test_dir2/rename_dir2/", "/test_dir/test_dir2/rename_dir3"))
			dispatcher.call(FuncRename, "/test_dir/test_dir2/rename_dir2/:/test_dir/test_dir2/rename_dir3", t)

			fi, err = vfs.Stat("/test_dir/test_dir2/rename_dir3")
			assert.NoError(t, err)
			if fi != nil {
				assert.Equal(t, fi.Name(), "rename_dir3")
				assert.True(t, fi.IsDir())
			}

			_, err = vfs.Stat("/test_dir/test_dir2/rename_dir2/")
			assert.Error(t, err)

			assert.NoError(t, vfs.RemoveAll("/test_dir/test_dir2/rename_dir3"))
		})

		t.Run("test Rename error", func(t *testing.T) {
			assert.Error(t, vfs.Rename("/noexist/noexist.txt", "/test_dir/noexist.txt"))
			dispatcher.call(FuncRenameErr, "/noexist/noexist.txt:/test_dir/noexist.txt", t)

			_, err = vfs.Create("/test_dir/rename.txt")
			assert.NoError(t, err)

			assert.Error(t, vfs.Rename("/test_dir/rename.txt", "/noexist/noexist.txt"))
			dispatcher.call(FuncRenameErr, "/test_dir/rename.txt:/noexist/noexist.txt", t)

			assert.NoError(t, vfs.Remove("/test_dir/rename.txt"))
		})

	})

	t.Run("test Sub", func(t *testing.T) {
		subVfs, err := vfs.Sub("/test_dir")
		assert.NoError(t, err)

		assert.False(t, subVfs.IsFile("test.txt"))
		assert.True(t, subVfs.IsFile("test1.txt"))

		dispatcher.call(FuncSub, "/test_dir", t)

		_, err = vfs.Sub("/noexist")
		assert.Error(t, err)
		dispatcher.call(FuncSubErr, "/noexist", t)

		_, err = vfs.Sub(".")
		assert.Error(t, err)
		dispatcher.call(FuncSubErr, ".", t)

		_, err = vfs.Sub("..")
		assert.Error(t, err)
		dispatcher.call(FuncSubErr, "..", t)
	})

	t.Run("test ReadDir", func(t *testing.T) {
		dirs, err := filesystem.ReadDir(vfs, "/")
		assert.NoError(t, err)

		assert.Len(t, dirs, 2)

		for _, dir := range dirs {
			if dir.IsDir() {
				assert.Equal(t, dir.Name(), "test_dir")
			} else {
				assert.Equal(t, dir.Name(), "test.txt")
			}

			assert.NotNil(t, dir.Type())
			fi, err := dir.Info()
			assert.NotNil(t, fi)
			assert.NoError(t, err)
		}

		dirs, err = filesystem.ReadDir(vfs, "/noexist")
		assert.Error(t, err)
		assert.Len(t, dirs, 0)

	})

	t.Run("test OpenFile", func(t *testing.T) {
		var (
			err error
		)
		_, err = filesystem.OpenFile(vfs, "test.txt")
		assert.NoError(t, err)

		_, err = filesystem.OpenFile(vfs, "noexist")
		assert.Error(t, err)
	})

	t.Run("test WriteFile and ReadFile", func(t *testing.T) {
		buf := bytes.NewBufferString("hello world")
		assert.NoError(t, filesystem.WriteFile(vfs, "test.txt", buf.Bytes()))

		body, err := filesystem.ReadFile(vfs, "test.txt")

		assert.NoError(t, err)

		assert.Equal(t, string(body), buf.String())
	})

	t.Run("clear", func(t *testing.T) {
		assert.NoError(t, vfs.RemoveAll("/"))
	})

}
