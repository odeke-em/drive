package binding

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-martini/martini"
)

var fileTestCases = []fileTestCase{
	{
		description: "Single file",
		singleFile: &fileInfo{
			fileName: "message.txt",
			data:     "All your binding are belong to us",
		},
	},
	{
		description: "Multiple files",
		multipleFiles: []*fileInfo{
			&fileInfo{
				fileName: "cool-gopher-fact.txt",
				data:     "Did you know? https://plus.google.com/+MatthewHolt/posts/GmVfd6TPJ51",
			},
			&fileInfo{
				fileName: "gophercon2014.txt",
				data:     "@bradfitz has a Go time machine: https://twitter.com/mholt6/status/459463953395875840",
			},
		},
	},
	{
		description: "Single file and multiple files",
		singleFile: &fileInfo{
			fileName: "social media.txt",
			data:     "Hey, you should follow @mholt6 (Twitter) or +MatthewHolt (Google+)",
		},
		multipleFiles: []*fileInfo{
			&fileInfo{
				fileName: "thank you!",
				data:     "Also, thanks to all the contributors of this package!",
			},
			&fileInfo{
				fileName: "btw...",
				data:     "This tool translates JSON into Go structs: http://mholt.github.io/json-to-go/",
			},
		},
	},
}

func TestFileUploads(t *testing.T) {
	for _, testCase := range fileTestCases {
		performFileTest(t, MultipartForm, testCase)
	}
}

func performFileTest(t *testing.T, binder handlerFunc, testCase fileTestCase) {
	httpRecorder := httptest.NewRecorder()
	m := martini.Classic()

	fileTestHandler := func(actual BlogPost, errs Errors) {
		assertFileAsExpected(t, testCase, actual.HeaderImage, testCase.singleFile)

		if len(testCase.multipleFiles) != len(actual.Pictures) {
			t.Errorf("For '%s': Expected %d multiple files, but actually had %d instead",
				testCase.description, len(testCase.multipleFiles), len(actual.Pictures))
		}

		for i, expectedFile := range testCase.multipleFiles {
			if i >= len(actual.Pictures) {
				break
			}
			assertFileAsExpected(t, testCase, actual.Pictures[i], expectedFile)
		}
	}

	m.Post(testRoute, binder(BlogPost{}), func(actual BlogPost, errs Errors) {
		fileTestHandler(actual, errs)
	})

	m.ServeHTTP(httpRecorder, buildRequestWithFile(testCase))

	switch httpRecorder.Code {
	case http.StatusNotFound:
		panic("Routing is messed up in test fixture (got 404): check methods and paths")
	case http.StatusInternalServerError:
		panic("Something bad happened on '" + testCase.description + "'")
	}
}

func assertFileAsExpected(t *testing.T, testCase fileTestCase, actual *multipart.FileHeader, expected *fileInfo) {
	if expected == nil && actual == nil {
		return
	}

	if expected != nil && actual == nil {
		t.Errorf("For '%s': Expected to have a file, but didn't",
			testCase.description)
		return
	} else if expected == nil && actual != nil {
		t.Errorf("For '%s': Did not expect a file, but ended up having one!",
			testCase.description)
		return
	}

	if actual.Filename != expected.fileName {
		t.Errorf("For '%s': expected file name to be '%s' but got '%s'",
			testCase.description, expected.fileName, actual.Filename)
	}

	actualMultipleFileData := unpackFileHeaderData(actual)

	if actualMultipleFileData != expected.data {
		t.Errorf("For '%s': expected file data to be '%s' but got '%s'",
			testCase.description, expected.data, actualMultipleFileData)
	}
}

func buildRequestWithFile(testCase fileTestCase) *http.Request {
	b := &bytes.Buffer{}
	w := multipart.NewWriter(b)

	if testCase.singleFile != nil {
		formFileSingle, err := w.CreateFormFile("headerImage", testCase.singleFile.fileName)
		if err != nil {
			panic("Could not create FormFile (single file): " + err.Error())
		}
		formFileSingle.Write([]byte(testCase.singleFile.data))
	}

	for _, file := range testCase.multipleFiles {
		formFileMultiple, err := w.CreateFormFile("picture", file.fileName)
		if err != nil {
			panic("Could not create FormFile (multiple files): " + err.Error())
		}
		formFileMultiple.Write([]byte(file.data))
	}

	err := w.Close()
	if err != nil {
		panic("Could not close multipart writer: " + err.Error())
	}

	req, err := http.NewRequest("POST", testRoute, b)
	if err != nil {
		panic("Could not create file upload request: " + err.Error())
	}

	req.Header.Set("Content-Type", w.FormDataContentType())

	return req
}

func unpackFileHeaderData(fh *multipart.FileHeader) string {
	if fh == nil {
		return ""
	}

	f, err := fh.Open()
	if err != nil {
		panic("Could not open file header:" + err.Error())
	}
	defer f.Close()

	var fb bytes.Buffer
	_, err = fb.ReadFrom(f)
	if err != nil {
		panic("Could not read from file header:" + err.Error())
	}

	return fb.String()
}

type (
	fileTestCase struct {
		description   string
		input         BlogPost
		singleFile    *fileInfo
		multipleFiles []*fileInfo
	}

	multipartFileFormTestCase struct {
		description   string
		fields        map[string]string
		multipleFiles []*fileInfo
	}

	fileInfo struct {
		fileName string
		data     string
	}
)

var multipartFileFormTestCases = []multipartFileFormTestCase{
	{
		description: "Test upload files",
		multipleFiles: []*fileInfo{
			&fileInfo{
				fileName: "cool-gopher-fact.txt",
				data:     "Did you know? https://plus.google.com/+MatthewHolt/posts/GmVfd6TPJ51",
			},
			&fileInfo{
				fileName: "gophercon2014.txt",
				data:     "@bradfitz has a Go time machine: https://twitter.com/mholt6/status/459463953395875840",
			},
		},
	},
	{
		description: "Test upload files with empty input[type=file]",
		multipleFiles: []*fileInfo{
			&fileInfo{
				fileName: "cool-gopher-fact.txt",
				data:     "Did you know? https://plus.google.com/+MatthewHolt/posts/GmVfd6TPJ51",
			},
			&fileInfo{
				fileName: "gophercon2014.txt",
				data:     "@bradfitz has a Go time machine: https://twitter.com/mholt6/status/459463953395875840",
			},
			&fileInfo{},
		},
	},
	{
		description: "Test send post data",
		fields: map[string]string{
			"test": "data",
			"save": "",
		},
	},
	{
		description: "Test send post data and upload files with empty input[type=file]",
		fields: map[string]string{
			"test": "data",
			"save": "",
		},
		multipleFiles: []*fileInfo{
			&fileInfo{
				fileName: "cool-gopher-fact.txt",
				data:     "Did you know? https://plus.google.com/+MatthewHolt/posts/GmVfd6TPJ51",
			},
			&fileInfo{
				fileName: "gophercon2014.txt",
				data:     "@bradfitz has a Go time machine: https://twitter.com/mholt6/status/459463953395875840",
			},
			&fileInfo{},
			&fileInfo{},
		},
	},
}

func TestMultipartFilesForm(t *testing.T) {
	for _, testCase := range multipartFileFormTestCases {
		performMultipartFilesFormTest(t, MultipartForm, testCase)
	}
}

func performMultipartFilesFormTest(t *testing.T, binder handlerFunc, testCase multipartFileFormTestCase) {
	httpRecorder := httptest.NewRecorder()
	m := martini.Classic()

	m.Post(testRoute, binder(BlogPost{}), func(actual BlogPost, errs Errors) {
		expectedCountFiles := 0
		actualCountFiles := 0

		for _, value := range testCase.multipleFiles {
			if value.fileName != "" {
				expectedCountFiles++ // Counting only not null files
			}
		}

		for _, value := range actual.Pictures {
			if value != nil {
				actualCountFiles++ // Counting only not null files
			}
		}

		if actualCountFiles != expectedCountFiles {
			t.Errorf("'%s': expected\n'%d'\nbut got\n'%d'", testCase.description, expectedCountFiles, actualCountFiles)
		}
	})

	multipartPayload, mpWriter := makeMultipartFilesPayload(testCase)

	req, err := http.NewRequest("POST", testRoute, multipartPayload)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Content-Type", mpWriter.FormDataContentType())

	err = mpWriter.Close()
	if err != nil {
		panic(err)
	}

	m.ServeHTTP(httpRecorder, req)

	switch httpRecorder.Code {
	case http.StatusNotFound:
		panic("Routing is messed up in test fixture (got 404): check methods and paths")
	case http.StatusInternalServerError:
		panic("Something bad happened on '" + testCase.description + "'")
	}
}

// Writes the input from a test case into a buffer using the multipart writer.
func makeMultipartFilesPayload(testCase multipartFileFormTestCase) (*bytes.Buffer, *multipart.Writer) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	var boundary string

	for key, value := range testCase.fields {
		boundary += "--" + writer.Boundary() + "\nContent-Disposition: form-data; name=\"" + key + "\";\n\n" + value + "\n\n"
	}

	for _, file := range testCase.multipleFiles {
		boundary += "--" + writer.Boundary() + "\nContent-Disposition: form-data; name=\"picture\"; filename=\"" + file.fileName + "\"\nContent-Type: text/plain\n\n" + file.data + "\n\n"
	}

	boundary += "--" + writer.Boundary() + "--\n"

	body.Write([]byte(boundary))
	return body, writer
}
