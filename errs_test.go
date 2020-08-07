package errs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"syscall"
	"testing"
)

type testError struct {
	Msg string
	Err error
}

func (t *testError) Error() string {
	return strings.Join([]string{t.Msg, t.Err.Error()}, ": ")
}
func (t *testError) Unwrap() error {
	return t.Err
}
func (t *testError) MarshalJSON() ([]byte, error) {
	if t == nil {
		return []byte("null"), nil
	}
	elms := []string{}
	elms = append(elms, fmt.Sprintf(`"Type":%q`, fmt.Sprintf("%T", t)))
	msgBuf := &bytes.Buffer{}
	json.HTMLEscape(msgBuf, []byte(fmt.Sprintf(`"Msg":%q`, t.Error())))
	elms = append(elms, msgBuf.String())
	if t.Err != nil && !reflect.ValueOf(t.Err).IsZero() {
		elms = append(elms, fmt.Sprintf(`"Err":%s`, EncodeJSON(t.Err)))
	}

	return []byte("{" + strings.Join(elms, ",") + "}"), nil
}

var (
	nilErr         = New("") // nil object
	nilValueErr    = (*Error)(nil)
	errTest        = New("\"Error\" for test")
	wrapedErrTest  = Wrap(errTest)
	wrapedErrTest2 = &testError{Msg: "test for testError", Err: wrapedErrTest}
)

func TestNil(t *testing.T) {
	testCases := []struct {
		err     error
		typeStr string
		ptr     string
		msg     string
		detail  string
		json    string
		badStr  string
	}{
		{
			err:     nilErr,
			typeStr: "<nil>",
			ptr:     "%!p(<nil>)",
			msg:     "<nil>",
			detail:  `<nil>`,
			json:    `<nil>`,
			badStr:  `%!d(<nil>)`,
		},
		{
			err:     nilValueErr,
			typeStr: "*errs.Error",
			ptr:     "0x0",
			msg:     "<nil>",
			detail:  `<nil>`,
			json:    `null`,
			badStr:  `%!d(<nil>)`,
		},
	}

	for _, tc := range testCases {
		str := fmt.Sprintf("%T", tc.err)
		if str != tc.typeStr {
			t.Errorf("Type of Wrap(\"%v\") is %v, want %v", tc.err, str, tc.typeStr)
		}
		str = fmt.Sprintf("%p", tc.err)
		if str != tc.ptr {
			t.Errorf("Pointer of Wrap(\"%v\") is %v, want %v", tc.err, str, tc.ptr)
		}
		str = fmt.Sprintf("%v", tc.err)
		if str != tc.msg {
			t.Errorf("Wrap(\"%v\") is %v, want %v", tc.err, str, tc.msg)
		}
		str = fmt.Sprintf("%#v", tc.err)
		if str != tc.detail {
			t.Errorf("Wrap(\"%v\") is %v, want %v", tc.err, str, tc.detail)
		}
		str = fmt.Sprintf("%+v", tc.err)
		if str != tc.json {
			t.Errorf("Wrap(\"%v\") is %v, want %v", tc.err, str, tc.json)
		}
		str = fmt.Sprintf("%d", tc.err)
		if str != tc.badStr {
			t.Errorf("Wrap(\"%v\") is %v, want %v", tc.err, str, tc.badStr)
		}
	}
}

func TestNewWithCause(t *testing.T) {
	testCases := []struct {
		err     error
		typeStr string
		ptr     string
		msg     string
		detail  string
		json    string
		badStr  string
	}{
		{
			err:     nilErr,
			typeStr: "*errs.Error",
			ptr:     "0x0",
			msg:     "wrapped message",
			detail:  `*errs.Error{Err:&errors.errorString{s:"wrapped message"}, Cause:<nil>, Context:map[string]interface {}{"foo":"bar", "function":"github.com/spiegel-im-spiegel/errs.TestNewWithCause", "num":1}}`,
			json:    `{"Type":"*errs.Error","Err":{"Type":"*errors.errorString","Msg":"wrapped message"},"Context":{"foo":"bar","function":"github.com/spiegel-im-spiegel/errs.TestNewWithCause","num":1}}`,
			badStr:  `%!d(*errs.Error{Err:&errors.errorString{s:"wrapped message"}, Cause:<nil>, Context:map[string]interface {}{"foo":"bar", "function":"github.com/spiegel-im-spiegel/errs.TestNewWithCause", "num":1}})`,
		},
		{
			err:     nilValueErr,
			typeStr: "*errs.Error",
			ptr:     "0x0",
			msg:     "wrapped message: <nil>",
			detail:  `*errs.Error{Err:&errors.errorString{s:"wrapped message"}, Cause:<nil>, Context:map[string]interface {}{"foo":"bar", "function":"github.com/spiegel-im-spiegel/errs.TestNewWithCause", "num":1}}`,
			json:    `{"Type":"*errs.Error","Err":{"Type":"*errors.errorString","Msg":"wrapped message"},"Context":{"foo":"bar","function":"github.com/spiegel-im-spiegel/errs.TestNewWithCause","num":1}}`,
			badStr:  `%!d(*errs.Error{Err:&errors.errorString{s:"wrapped message"}, Cause:<nil>, Context:map[string]interface {}{"foo":"bar", "function":"github.com/spiegel-im-spiegel/errs.TestNewWithCause", "num":1}})`,
		},
		{
			err:     os.ErrInvalid,
			typeStr: "*errs.Error",
			ptr:     "0x0",
			msg:     "wrapped message: invalid argument",
			detail:  `*errs.Error{Err:&errors.errorString{s:"wrapped message"}, Cause:&errors.errorString{s:"invalid argument"}, Context:map[string]interface {}{"foo":"bar", "function":"github.com/spiegel-im-spiegel/errs.TestNewWithCause", "num":1}}`,
			json:    `{"Type":"*errs.Error","Err":{"Type":"*errors.errorString","Msg":"wrapped message"},"Context":{"foo":"bar","function":"github.com/spiegel-im-spiegel/errs.TestNewWithCause","num":1},"Cause":{"Type":"*errors.errorString","Msg":"invalid argument"}}`,
			badStr:  `%!d(*errs.Error{Err:&errors.errorString{s:"wrapped message"}, Cause:&errors.errorString{s:"invalid argument"}, Context:map[string]interface {}{"foo":"bar", "function":"github.com/spiegel-im-spiegel/errs.TestNewWithCause", "num":1}})`,
		},
		{
			err:     errTest,
			typeStr: "*errs.Error",
			ptr:     "0x0",
			msg:     "wrapped message: \"Error\" for test",
			detail:  `*errs.Error{Err:&errors.errorString{s:"wrapped message"}, Cause:*errs.Error{Err:&errors.errorString{s:"\"Error\" for test"}, Cause:<nil>, Context:map[string]interface {}{"function":"github.com/spiegel-im-spiegel/errs.init"}}, Context:map[string]interface {}{"foo":"bar", "function":"github.com/spiegel-im-spiegel/errs.TestNewWithCause", "num":1}}`,
			json:    `{"Type":"*errs.Error","Err":{"Type":"*errors.errorString","Msg":"wrapped message"},"Context":{"foo":"bar","function":"github.com/spiegel-im-spiegel/errs.TestNewWithCause","num":1},"Cause":{"Type":"*errs.Error","Err":{"Type":"*errors.errorString","Msg":"\"Error\" for test"},"Context":{"function":"github.com/spiegel-im-spiegel/errs.init"}}}`,
			badStr:  `%!d(*errs.Error{Err:&errors.errorString{s:"wrapped message"}, Cause:*errs.Error{Err:&errors.errorString{s:"\"Error\" for test"}, Cause:<nil>, Context:map[string]interface {}{"function":"github.com/spiegel-im-spiegel/errs.init"}}, Context:map[string]interface {}{"foo":"bar", "function":"github.com/spiegel-im-spiegel/errs.TestNewWithCause", "num":1}})`,
		},
		{
			err:     wrapedErrTest,
			typeStr: "*errs.Error",
			ptr:     "0x0",
			msg:     "wrapped message: \"Error\" for test",
			detail:  `*errs.Error{Err:&errors.errorString{s:"wrapped message"}, Cause:*errs.Error{Err:*errs.Error{Err:&errors.errorString{s:"\"Error\" for test"}, Cause:<nil>, Context:map[string]interface {}{"function":"github.com/spiegel-im-spiegel/errs.init"}}, Cause:<nil>, Context:map[string]interface {}{"function":"github.com/spiegel-im-spiegel/errs.init"}}, Context:map[string]interface {}{"foo":"bar", "function":"github.com/spiegel-im-spiegel/errs.TestNewWithCause", "num":1}}`,
			json:    `{"Type":"*errs.Error","Err":{"Type":"*errors.errorString","Msg":"wrapped message"},"Context":{"foo":"bar","function":"github.com/spiegel-im-spiegel/errs.TestNewWithCause","num":1},"Cause":{"Type":"*errs.Error","Err":{"Type":"*errs.Error","Err":{"Type":"*errors.errorString","Msg":"\"Error\" for test"},"Context":{"function":"github.com/spiegel-im-spiegel/errs.init"}},"Context":{"function":"github.com/spiegel-im-spiegel/errs.init"}}}`,
			badStr:  `%!d(*errs.Error{Err:&errors.errorString{s:"wrapped message"}, Cause:*errs.Error{Err:*errs.Error{Err:&errors.errorString{s:"\"Error\" for test"}, Cause:<nil>, Context:map[string]interface {}{"function":"github.com/spiegel-im-spiegel/errs.init"}}, Cause:<nil>, Context:map[string]interface {}{"function":"github.com/spiegel-im-spiegel/errs.init"}}, Context:map[string]interface {}{"foo":"bar", "function":"github.com/spiegel-im-spiegel/errs.TestNewWithCause", "num":1}})`,
		},
		{
			err:     wrapedErrTest2,
			typeStr: "*errs.Error",
			ptr:     "0x0",
			msg:     "wrapped message: test for testError: \"Error\" for test",
			detail:  `*errs.Error{Err:&errors.errorString{s:"wrapped message"}, Cause:&errs.testError{Msg:"test for testError", Err:*errs.Error{Err:*errs.Error{Err:&errors.errorString{s:"\"Error\" for test"}, Cause:<nil>, Context:map[string]interface {}{"function":"github.com/spiegel-im-spiegel/errs.init"}}, Cause:<nil>, Context:map[string]interface {}{"function":"github.com/spiegel-im-spiegel/errs.init"}}}, Context:map[string]interface {}{"foo":"bar", "function":"github.com/spiegel-im-spiegel/errs.TestNewWithCause", "num":1}}`,
			json:    `{"Type":"*errs.Error","Err":{"Type":"*errors.errorString","Msg":"wrapped message"},"Context":{"foo":"bar","function":"github.com/spiegel-im-spiegel/errs.TestNewWithCause","num":1},"Cause":{"Type":"*errs.testError","Msg":"test for testError: \"Error\" for test","Err":{"Type":"*errs.Error","Err":{"Type":"*errs.Error","Err":{"Type":"*errors.errorString","Msg":"\"Error\" for test"},"Context":{"function":"github.com/spiegel-im-spiegel/errs.init"}},"Context":{"function":"github.com/spiegel-im-spiegel/errs.init"}}}}`,
			badStr:  `%!d(*errs.Error{Err:&errors.errorString{s:"wrapped message"}, Cause:&errs.testError{Msg:"test for testError", Err:*errs.Error{Err:*errs.Error{Err:&errors.errorString{s:"\"Error\" for test"}, Cause:<nil>, Context:map[string]interface {}{"function":"github.com/spiegel-im-spiegel/errs.init"}}, Cause:<nil>, Context:map[string]interface {}{"function":"github.com/spiegel-im-spiegel/errs.init"}}}, Context:map[string]interface {}{"foo":"bar", "function":"github.com/spiegel-im-spiegel/errs.TestNewWithCause", "num":1}})`,
		},
	}

	for _, tc := range testCases {
		err := New("wrapped message", WithCause(tc.err), WithContext("foo", "bar"), WithContext("num", 1))
		str := fmt.Sprintf("%T", err)
		if str != tc.typeStr {
			t.Errorf("Type of Wrap(\"%v\") is %v, want %v", tc.err, str, tc.typeStr)
		}
		str = fmt.Sprintf("%p", err)
		if str == tc.ptr {
			t.Errorf("Pointer of Wrap(\"%v\") is %v, not want %v", tc.err, str, tc.ptr)
		} else {
			fmt.Println("Info:", str)
		}
		str = fmt.Sprintf("%v", err)
		if str != tc.msg {
			t.Errorf("Wrap(\"%v\") is %v, want %v", tc.err, str, tc.msg)
		}
		str = fmt.Sprintf("%#v", err)
		if str != tc.detail {
			t.Errorf("Wrap(\"%v\") is %v, want %v", tc.err, str, tc.detail)
		}
		str = fmt.Sprintf("%+v", err)
		if str != tc.json {
			t.Errorf("Wrap(\"%v\") is %v, want %v", tc.err, str, tc.json)
		}
		str = fmt.Sprintf("%d", err)
		if str != tc.badStr {
			t.Errorf("Wrap(\"%v\") is %v, want %v", tc.err, str, tc.badStr)
		}
		if err != nil {
			b, e := json.Marshal(err)
			if e != nil {
				t.Errorf("json.Marshal(\"%v\") is %v, want <nil>", tc.err, e)
			} else if string(b) != tc.json {
				t.Errorf("Wrap(\"%v\") is %v, want %v", tc.err, string(b), tc.json)
			}
			str = EncodeJSON(err)
			if str != tc.json {
				t.Errorf("Wrap(\"%v\") is %v, want %v", tc.err, str, tc.json)
			}
		}
	}
}

func TestWrapWithCause(t *testing.T) {
	testCases := []struct {
		err     error
		typeStr string
		ptr     string
		msg     string
		detail  string
		json    string
		badStr  string
	}{
		{
			err:     nilErr,
			typeStr: "*errs.Error",
			ptr:     "0x0",
			msg:     "wrapped message",
			detail:  `*errs.Error{Err:&errors.errorString{s:"wrapped message"}, Cause:<nil>, Context:map[string]interface {}{"foo":"bar", "function":"github.com/spiegel-im-spiegel/errs.TestWrapWithCause", "num":1}}`,
			json:    `{"Type":"*errs.Error","Err":{"Type":"*errors.errorString","Msg":"wrapped message"},"Context":{"foo":"bar","function":"github.com/spiegel-im-spiegel/errs.TestWrapWithCause","num":1}}`,
			badStr:  `%!d(*errs.Error{Err:&errors.errorString{s:"wrapped message"}, Cause:<nil>, Context:map[string]interface {}{"foo":"bar", "function":"github.com/spiegel-im-spiegel/errs.TestWrapWithCause", "num":1}})`,
		},
		{
			err:     nilValueErr,
			typeStr: "*errs.Error",
			ptr:     "0x0",
			msg:     "wrapped message: <nil>",
			detail:  `*errs.Error{Err:&errors.errorString{s:"wrapped message"}, Cause:<nil>, Context:map[string]interface {}{"foo":"bar", "function":"github.com/spiegel-im-spiegel/errs.TestWrapWithCause", "num":1}}`,
			json:    `{"Type":"*errs.Error","Err":{"Type":"*errors.errorString","Msg":"wrapped message"},"Context":{"foo":"bar","function":"github.com/spiegel-im-spiegel/errs.TestWrapWithCause","num":1}}`,
			badStr:  `%!d(*errs.Error{Err:&errors.errorString{s:"wrapped message"}, Cause:<nil>, Context:map[string]interface {}{"foo":"bar", "function":"github.com/spiegel-im-spiegel/errs.TestWrapWithCause", "num":1}})`,
		},
		{
			err:     os.ErrInvalid,
			typeStr: "*errs.Error",
			ptr:     "0x0",
			msg:     "wrapped message: invalid argument",
			detail:  `*errs.Error{Err:&errors.errorString{s:"wrapped message"}, Cause:&errors.errorString{s:"invalid argument"}, Context:map[string]interface {}{"foo":"bar", "function":"github.com/spiegel-im-spiegel/errs.TestWrapWithCause", "num":1}}`,
			json:    `{"Type":"*errs.Error","Err":{"Type":"*errors.errorString","Msg":"wrapped message"},"Context":{"foo":"bar","function":"github.com/spiegel-im-spiegel/errs.TestWrapWithCause","num":1},"Cause":{"Type":"*errors.errorString","Msg":"invalid argument"}}`,
			badStr:  `%!d(*errs.Error{Err:&errors.errorString{s:"wrapped message"}, Cause:&errors.errorString{s:"invalid argument"}, Context:map[string]interface {}{"foo":"bar", "function":"github.com/spiegel-im-spiegel/errs.TestWrapWithCause", "num":1}})`,
		},
		{
			err:     errTest,
			typeStr: "*errs.Error",
			ptr:     "0x0",
			msg:     "wrapped message: \"Error\" for test",
			detail:  `*errs.Error{Err:&errors.errorString{s:"wrapped message"}, Cause:*errs.Error{Err:&errors.errorString{s:"\"Error\" for test"}, Cause:<nil>, Context:map[string]interface {}{"function":"github.com/spiegel-im-spiegel/errs.init"}}, Context:map[string]interface {}{"foo":"bar", "function":"github.com/spiegel-im-spiegel/errs.TestWrapWithCause", "num":1}}`,
			json:    `{"Type":"*errs.Error","Err":{"Type":"*errors.errorString","Msg":"wrapped message"},"Context":{"foo":"bar","function":"github.com/spiegel-im-spiegel/errs.TestWrapWithCause","num":1},"Cause":{"Type":"*errs.Error","Err":{"Type":"*errors.errorString","Msg":"\"Error\" for test"},"Context":{"function":"github.com/spiegel-im-spiegel/errs.init"}}}`,
			badStr:  `%!d(*errs.Error{Err:&errors.errorString{s:"wrapped message"}, Cause:*errs.Error{Err:&errors.errorString{s:"\"Error\" for test"}, Cause:<nil>, Context:map[string]interface {}{"function":"github.com/spiegel-im-spiegel/errs.init"}}, Context:map[string]interface {}{"foo":"bar", "function":"github.com/spiegel-im-spiegel/errs.TestWrapWithCause", "num":1}})`,
		},
		{
			err:     wrapedErrTest,
			typeStr: "*errs.Error",
			ptr:     "0x0",
			msg:     "wrapped message: \"Error\" for test",
			detail:  `*errs.Error{Err:&errors.errorString{s:"wrapped message"}, Cause:*errs.Error{Err:*errs.Error{Err:&errors.errorString{s:"\"Error\" for test"}, Cause:<nil>, Context:map[string]interface {}{"function":"github.com/spiegel-im-spiegel/errs.init"}}, Cause:<nil>, Context:map[string]interface {}{"function":"github.com/spiegel-im-spiegel/errs.init"}}, Context:map[string]interface {}{"foo":"bar", "function":"github.com/spiegel-im-spiegel/errs.TestWrapWithCause", "num":1}}`,
			json:    `{"Type":"*errs.Error","Err":{"Type":"*errors.errorString","Msg":"wrapped message"},"Context":{"foo":"bar","function":"github.com/spiegel-im-spiegel/errs.TestWrapWithCause","num":1},"Cause":{"Type":"*errs.Error","Err":{"Type":"*errs.Error","Err":{"Type":"*errors.errorString","Msg":"\"Error\" for test"},"Context":{"function":"github.com/spiegel-im-spiegel/errs.init"}},"Context":{"function":"github.com/spiegel-im-spiegel/errs.init"}}}`,
			badStr:  `%!d(*errs.Error{Err:&errors.errorString{s:"wrapped message"}, Cause:*errs.Error{Err:*errs.Error{Err:&errors.errorString{s:"\"Error\" for test"}, Cause:<nil>, Context:map[string]interface {}{"function":"github.com/spiegel-im-spiegel/errs.init"}}, Cause:<nil>, Context:map[string]interface {}{"function":"github.com/spiegel-im-spiegel/errs.init"}}, Context:map[string]interface {}{"foo":"bar", "function":"github.com/spiegel-im-spiegel/errs.TestWrapWithCause", "num":1}})`,
		},
		{
			err:     wrapedErrTest2,
			typeStr: "*errs.Error",
			ptr:     "0x0",
			msg:     "wrapped message: test for testError: \"Error\" for test",
			detail:  `*errs.Error{Err:&errors.errorString{s:"wrapped message"}, Cause:&errs.testError{Msg:"test for testError", Err:*errs.Error{Err:*errs.Error{Err:&errors.errorString{s:"\"Error\" for test"}, Cause:<nil>, Context:map[string]interface {}{"function":"github.com/spiegel-im-spiegel/errs.init"}}, Cause:<nil>, Context:map[string]interface {}{"function":"github.com/spiegel-im-spiegel/errs.init"}}}, Context:map[string]interface {}{"foo":"bar", "function":"github.com/spiegel-im-spiegel/errs.TestWrapWithCause", "num":1}}`,
			json:    `{"Type":"*errs.Error","Err":{"Type":"*errors.errorString","Msg":"wrapped message"},"Context":{"foo":"bar","function":"github.com/spiegel-im-spiegel/errs.TestWrapWithCause","num":1},"Cause":{"Type":"*errs.testError","Msg":"test for testError: \"Error\" for test","Err":{"Type":"*errs.Error","Err":{"Type":"*errs.Error","Err":{"Type":"*errors.errorString","Msg":"\"Error\" for test"},"Context":{"function":"github.com/spiegel-im-spiegel/errs.init"}},"Context":{"function":"github.com/spiegel-im-spiegel/errs.init"}}}}`,
			badStr:  `%!d(*errs.Error{Err:&errors.errorString{s:"wrapped message"}, Cause:&errs.testError{Msg:"test for testError", Err:*errs.Error{Err:*errs.Error{Err:&errors.errorString{s:"\"Error\" for test"}, Cause:<nil>, Context:map[string]interface {}{"function":"github.com/spiegel-im-spiegel/errs.init"}}, Cause:<nil>, Context:map[string]interface {}{"function":"github.com/spiegel-im-spiegel/errs.init"}}}, Context:map[string]interface {}{"foo":"bar", "function":"github.com/spiegel-im-spiegel/errs.TestWrapWithCause", "num":1}})`,
		},
	}

	for _, tc := range testCases {
		err := Wrap(errors.New("wrapped message"), WithCause(tc.err), WithContext("foo", "bar"), WithContext("num", 1))
		str := fmt.Sprintf("%T", err)
		if str != tc.typeStr {
			t.Errorf("Type of Wrap(\"%v\") is %v, want %v", tc.err, str, tc.typeStr)
		}
		str = fmt.Sprintf("%p", err)
		if str == tc.ptr {
			t.Errorf("Pointer of Wrap(\"%v\") is %v, not want %v", tc.err, str, tc.ptr)
		} else {
			fmt.Println("Info:", str)
		}
		str = fmt.Sprintf("%v", err)
		if str != tc.msg {
			t.Errorf("Wrap(\"%v\") is %v, want %v", tc.err, str, tc.msg)
		}
		str = fmt.Sprintf("%#v", err)
		if str != tc.detail {
			t.Errorf("Wrap(\"%v\") is %v, want %v", tc.err, str, tc.detail)
		}
		str = fmt.Sprintf("%+v", err)
		if str != tc.json {
			t.Errorf("Wrap(\"%v\") is %v, want %v", tc.err, str, tc.json)
		}
		str = fmt.Sprintf("%d", err)
		if str != tc.badStr {
			t.Errorf("Wrap(\"%v\") is %v, want %v", tc.err, str, tc.badStr)
		}
		if err != nil {
			b, e := json.Marshal(err)
			if e != nil {
				t.Errorf("json.Marshal(\"%v\") is %v, want <nil>", tc.err, e)
			} else if string(b) != tc.json {
				t.Errorf("Wrap(\"%v\") is %v, want %v", tc.err, string(b), tc.json)
			}
			str = EncodeJSON(err)
			if str != tc.json {
				t.Errorf("Wrap(\"%v\") is %v, want %v", tc.err, str, tc.json)
			}
		}
	}
}

func TestCause(t *testing.T) {
	testCases := []struct {
		err   error
		cause error
	}{
		{err: nil, cause: nil},
		{err: os.ErrInvalid, cause: os.ErrInvalid},
		{err: New("wrapped error", WithCause(os.ErrInvalid)), cause: os.ErrInvalid},
	}

	for _, tc := range testCases {
		c := Cause(tc.err)
		if c != tc.cause {
			t.Errorf("result Cause(\"%v\") is \"%v\", want %v", tc.err, c, tc.cause)
		}
	}
}

func TestIs(t *testing.T) {
	testCases := []struct {
		err    error
		res    bool
		target error
	}{
		{err: nil, res: true, target: nil},
		{err: New("error"), res: false, target: nil},
		{err: New(""), res: true, target: nil},
		{err: Wrap(nil), res: true, target: nil},
		{err: nil, res: false, target: errTest},
		{err: errTest, res: false, target: nil},
		{err: errTest, res: true, target: errTest},
		{err: errTest, res: false, target: os.ErrInvalid},
		{err: New("wrapped error", WithCause(os.ErrInvalid)), res: true, target: os.ErrInvalid},
		{err: New("wrapped error", WithCause(os.ErrInvalid)), res: false, target: errTest},
		{err: New("wrapped error", WithCause(errTest)), res: true, target: errTest},
		{err: New("wrapped error", WithCause(errTest)), res: true, target: wrapedErrTest},
		{err: New("wrapped error", WithCause(errTest)), res: false, target: os.ErrInvalid},
	}

	for _, tc := range testCases {
		if ok := Is(tc.err, tc.target); ok != tc.res {
			t.Errorf("result Is(\"%v\", \"%v\") is %v, want %v", tc.err, tc.target, ok, tc.res)
		}
	}
}

func TestAs(t *testing.T) {
	testCases := []struct {
		err   error
		res   bool
		cause error
	}{
		{err: nil, res: false, cause: nil},
		{err: New("wrapped error", WithCause(syscall.ENOENT)), res: true, cause: syscall.ENOENT},
	}

	for _, tc := range testCases {
		var cs syscall.Errno
		if ok := As(tc.err, &cs); ok != tc.res {
			t.Errorf("result if As(\"%v\") is %v, want %v", tc.err, ok, tc.res)
			if ok && cs != tc.cause {
				t.Errorf("As(\"%v\") = \"%v\", want \"%v\"", tc.err, cs, tc.cause)
			}
		}
	}
}

func TestUnwrap(t *testing.T) {
	testCases := []struct {
		err   error
		cause error
	}{
		{err: nil, cause: nil},
		{err: syscall.ENOENT, cause: nil},
		{err: New("wrapped error", WithCause(syscall.ENOENT)), cause: syscall.ENOENT},
	}

	for _, tc := range testCases {
		cs := Unwrap(tc.err)
		if cs != tc.cause {
			t.Errorf("As(\"%v\") = \"%v\", want \"%v\"", tc.err, cs, tc.cause)
		}
	}
}

/* Copyright 2019,2020 Spiegel
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * 	http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
