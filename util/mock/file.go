package mock

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/calvine/filejitsu/util"
	"github.com/google/uuid"
)

type CleanupFunction func() error

type IncompleteMockFile struct {
	FileName     string
	RelativePath string
	Content      string
}

// type IncompleteContentMap map[string]IncompleteMockFile

type ContentMap map[string]MockFile

type MockFile struct {
	Content  string
	FullPath string
	Hash     string
}

func getMockDirContentMap(rootDir string) (ContentMap, error) {
	content := []IncompleteMockFile{
		{
			FileName:     "file1.txt",
			Content:      "Magna nulla magna do qui irure.",
			RelativePath: filepath.Join("file1.txt"),
		},
		{
			FileName:     "file2.txt",
			Content:      "Ea incididunt elit elit duis eiusmod quis pariatur. Tempor irure do velit eu voluptate nisi cupidatat sit consequat aliqua velit.",
			RelativePath: filepath.Join("file2.txt"),
		},
		{
			FileName: "bigfile.txt",
			Content: `
			Exercitation exercitation elit laborum proident consectetur aliquip incididunt amet nulla exercitation reprehenderit ullamco cupidatat. Exercitation aliqua esse velit ad. Laborum quis irure nisi commodo qui cillum nisi consequat voluptate. Consectetur laborum culpa proident deserunt nulla quis minim.

Elit veniam et esse eu duis aliqua elit voluptate velit. Reprehenderit dolore laborum exercitation esse pariatur in et pariatur aliqua. Deserunt occaecat excepteur eu officia et sunt mollit eu incididunt sit laborum. Nulla in labore esse reprehenderit. Veniam do ad ipsum ea non non ad. Aliqua aute excepteur Lorem ad dolore quis cupidatat ut cupidatat dolor esse elit anim. Labore excepteur anim et exercitation exercitation consequat veniam reprehenderit occaecat esse.

Enim commodo ut anim cupidatat Lorem incididunt sit ipsum id amet deserunt. Deserunt duis occaecat id velit est fugiat pariatur. Anim dolore aute ad consequat aute sit nostrud duis ipsum. Nulla pariatur labore in incididunt fugiat. Est exercitation labore magna ullamco pariatur eu Lorem elit veniam elit. Ipsum in Lorem culpa et enim quis ipsum enim irure esse elit commodo exercitation. Consequat ad eu mollit laboris sit.

Duis officia mollit cillum consectetur pariatur ullamco. Ipsum do dolor amet duis ea exercitation fugiat dolor eu laboris elit et officia. Mollit irure aute ex ullamco velit ex et non cupidatat dolore commodo aute magna. Culpa in Lorem mollit nulla proident.

Proident irure qui nulla incididunt pariatur. Magna adipisicing ullamco non sit aute do cupidatat. Non consequat excepteur enim aliquip.

Nostrud id laboris quis eu consequat aliqua id adipisicing officia reprehenderit pariatur. Ex mollit sunt aliqua Lorem sunt. Fugiat sunt velit voluptate quis anim. Et cillum deserunt minim labore. In eiusmod eiusmod adipisicing ut aliqua dolor dolore duis labore. Sit amet exercitation id aliqua exercitation quis voluptate. Cillum est magna veniam quis ad aute esse sunt.

Cillum eu elit deserunt fugiat culpa est nostrud est sint quis tempor. Anim Lorem irure eiusmod do id est dolore ex minim ad. Minim incididunt amet nisi culpa. Mollit magna qui elit consequat incididunt. Est laboris minim nostrud laboris aute id proident. Aliqua eiusmod cupidatat Lorem adipisicing consequat mollit dolore do sint ea veniam anim proident enim.

Est ipsum sunt do irure id non culpa do ipsum voluptate. Lorem elit dolore aliqua anim proident do tempor adipisicing exercitation consectetur. Duis eu et elit commodo occaecat proident laboris eiusmod ea Lorem. Mollit nulla labore laborum aliquip incididunt pariatur culpa do qui. Dolor amet id esse velit minim tempor.

Lorem laboris enim sunt amet exercitation deserunt consequat incididunt est eu anim consequat. Enim ex eu commodo reprehenderit esse veniam sunt reprehenderit commodo dolor ea officia proident eu. Dolor id est ullamco exercitation occaecat officia dolore duis adipisicing quis deserunt. Sit fugiat cillum ea quis do dolor id. Esse dolore aute consectetur esse ex aliqua irure consectetur non officia culpa excepteur veniam proident.

Sit quis reprehenderit Lorem incididunt veniam ut sit mollit proident pariatur labore. Laboris culpa ullamco veniam qui officia. Qui adipisicing eiusmod sunt nostrud cillum mollit sit.
			`,
			RelativePath: filepath.Join("nested", "bigfile.txt"),
		},
		{
			FileName:     "file.txt",
			Content:      "a double nested file",
			RelativePath: filepath.Join("nested", "nexted2", "file.txt"),
		},
	}
	return MakeMockDirContentMap(rootDir, content)
}

func MakeMockDirContentMap(rootDir string, content []IncompleteMockFile) (ContentMap, error) {
	contentMap := make(ContentMap)
	// get file content hashes
	mockLogger := NewMockLogger()
	for _, k := range content {
		contentBuffer := bytes.NewBuffer([]byte(k.Content))
		hash, err := util.Sha512HashData(mockLogger, contentBuffer)
		if err != nil {
			return nil, fmt.Errorf("failed to hash file %s - %w", k, err)
		}
		mockFile := MockFile{
			Content:  k.Content,
			FullPath: filepath.Join(rootDir, k.RelativePath),
			Hash:     hash,
		}
		contentMap[k.RelativePath] = mockFile
	}

	return contentMap, nil
}

func MakeGenericMockDirTree() (string, ContentMap, CleanupFunction, error) {
	rootDir := GetRandomDirName()
	content, err := getMockDirContentMap(rootDir)
	if err != nil {
		return "", nil, nil, err
	}
	return CreateCustomMockDirTree(rootDir, content)
}

// GetRandomDirName returns a random path to a directory in your OSes temp directory. IT DOES NOT MAKE THE DIRECTORY
func GetRandomDirName() string {
	tmpDir := os.TempDir()
	dirName := uuid.New().String()
	rootDir := filepath.Join(tmpDir, dirName)
	return rootDir
}

func CreateCustomMockDirTree(rootDir string, content ContentMap) (string, ContentMap, CleanupFunction, error) {
	for k, v := range content {
		filePath := filepath.Dir(v.FullPath)
		if err := os.MkdirAll(filePath, 0766); err != nil {
			return "", nil, nil, fmt.Errorf("failed to make dir %s - %w", filePath, err)
		}
		f, err := os.Create(v.FullPath)
		if err != nil {
			return "", nil, nil, fmt.Errorf("failed to create file %s - %w", k, err)
		}
		contentLength := len(v.Content)
		bytesWritten, err := f.Write([]byte(v.Content))
		if err != nil {
			return "", nil, nil, fmt.Errorf("failed to write content to file %s - %w", k, err)
		}
		if bytesWritten != contentLength {
			return "", nil, nil, fmt.Errorf("bytes written (%d) not equal to content length (%d)", bytesWritten, contentLength)
		}
		if err := f.Close(); err != nil {
			return "", nil, nil, fmt.Errorf("failed to close file %s - %w", k, err)
		}
	}
	cleanup := func() error {
		err := os.RemoveAll(rootDir)
		return err
	}
	return rootDir, content, cleanup, nil
}

func ConfirmContentMapMatches(rootPath string, content ContentMap) error {
	contentCopy := make(ContentMap)
	for k, v := range content {
		contentCopy[k] = v
	}
	mockLogger := NewMockLogger()
	err := filepath.Walk(rootPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// if relativePath == "." {
		// 	return filepath.SkipDir
		// }

		if info.IsDir() {
			// content does not track dirs, so skipping
			return nil
		}

		fullPath := path
		relativePath, _ := filepath.Rel(rootPath, path)
		c, ok := contentCopy[relativePath]
		if !ok {
			return fmt.Errorf("file in untar path is not in mock content: %s", fullPath)
		}

		f, err := os.Open(fullPath)
		if err != nil {
			return fmt.Errorf("failed to open file %s", fullPath)
		}

		hash, err := util.Sha512HashData(mockLogger, f)

		if err != nil {
			return fmt.Errorf("failed to hash file %s", fullPath)
		}

		if err := f.Close(); err != nil {
			return fmt.Errorf("failed to close file %s", fullPath)
		}

		if c.Hash != hash {
			return fmt.Errorf("%s hash does not match expected hash: got - %s expected - %s", fullPath, hash, c.Hash)
		}

		delete(contentCopy, relativePath)

		return nil
	})
	if err != nil {
		return err
	}
	numContentRemaining := len(contentCopy)
	if numContentRemaining != 0 {
		return fmt.Errorf("items still left in content array: %v", contentCopy)
	}
	return nil
}
