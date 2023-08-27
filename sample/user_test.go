// A test that uses a mock.
package user_test

import (
	"fmt"
	"testing"

	"go.uber.org/mock/gomock"
	user "go.uber.org/mock/sample"
	"go.uber.org/mock/sample/imp1"
	imp_four "go.uber.org/mock/sample/imp4"
)

func TestRemember(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// NewMockIndex 生成一个新的Index 是mock的
	mockIndex := NewMockIndex(ctrl)
	// 通过EXCEPT接口设置方法期望输入，比如这里设置Put方法输入"a"和1
	mockIndex.EXPECT().Put("a", 1) // literals work
	// 要求输入b且另一个参数是2
	mockIndex.EXPECT().Put("b", gomock.Eq(2)) // matchers work too

	// 这个方法应该返回一个error值，我们应该通过Return方法设置返回值。
	// 但是我们没有，所以他会报错，那也就会返回error，正好符合预期
	// NillableRet returns error. Not declaring it should result in a nil return.
	mockIndex.EXPECT().NillableRet()
	// Calls that returns something assignable to the return type.
	boolc := make(chan bool)
	// In this case, "chan bool" is assignable to "chan<- bool".
	mockIndex.EXPECT().ConcreteRet().Return(boolc)
	// In this case, nil is assignable to "chan<- bool".
	mockIndex.EXPECT().ConcreteRet().Return(nil)

	// Should be able to place expectations on variadic methods.
	mockIndex.EXPECT().Ellip("%d", 0, 1, 1, 2, 3) // direct args
	tri := []any{1, 3, 6, 10, 15}
	mockIndex.EXPECT().Ellip("%d", tri...) // args from slice
	mockIndex.EXPECT().EllipOnly(gomock.Eq("arg"))

	user.Remember(mockIndex, []string{"a", "b"}, []any{1, 2})
	// Check the ConcreteRet calls.
	if c := mockIndex.ConcreteRet(); c != boolc {
		t.Errorf("ConcreteRet: got %v, want %v", c, boolc)
	}
	if c := mockIndex.ConcreteRet(); c != nil {
		t.Errorf("ConcreteRet: got %v, want nil", c)
	}

	// Try one with an action.
	calledString := ""
	// 这里将Put方法设置为无论传入什么参数都会将key记录到calledString里面。
	mockIndex.EXPECT().Put(gomock.Any(), gomock.Any()).Do(func(key string, _ any) {
		calledString = key
	})
	fmt.Println("calledString=", calledString)
	//
	mockIndex.EXPECT().NillableRet()
	user.Remember(mockIndex, []string{"blah"}, []any{7})
	if calledString != "blah" {
		t.Fatalf(`Uh oh. %q != "blah"`, calledString)
	}

	// Use Do with a nil arg.
	mockIndex.EXPECT().Put("nil-key", gomock.Any()).Do(func(key string, value any) {
		fmt.Println("GO!")
		if value != nil {
			t.Errorf("Put did not pass through nil; got %v", value)
		}
	})
	mockIndex.EXPECT().NillableRet()
	user.Remember(mockIndex, []string{"nil-key"}, []any{nil})
}

func TestVariadicFunction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIndex := NewMockIndex(ctrl)
	mockIndex.EXPECT().Ellip("%d", 5, 6, 7, 8).Do(func(format string, nums ...int) {
		sum := 0
		for _, value := range nums {
			sum += value
		}
		if sum != 26 {
			t.Errorf("Expected 26, got %d", sum)
		}
	})
	mockIndex.EXPECT().Ellip("%d", gomock.Any()).Do(func(format string, nums ...int) {
		sum := 0
		for _, value := range nums {
			sum += value
		}
		if sum != 10 {
			t.Errorf("Expected 10, got %d", sum)
		}
	})
	mockIndex.EXPECT().Ellip("%d", gomock.Any()).Do(func(format string, nums ...int) {
		sum := 0
		for _, value := range nums {
			sum += value
		}
		if sum != 0 {
			t.Errorf("Expected 0, got %d", sum)
		}
	})
	mockIndex.EXPECT().Ellip("%d", gomock.Any()).Do(func(format string, nums ...int) {
		sum := 0
		for _, value := range nums {
			sum += value
		}
		if sum != 0 {
			t.Errorf("Expected 0, got %d", sum)
		}
	})
	mockIndex.EXPECT().Ellip("%d").Do(func(format string, nums ...int) {
		sum := 0
		for _, value := range nums {
			sum += value
		}
		if sum != 0 {
			t.Errorf("Expected 0, got %d", sum)
		}
	})

	mockIndex.Ellip("%d", 1, 2, 3, 4) // Match second matcher.
	mockIndex.Ellip("%d", 5, 6, 7, 8) // Match first matcher.
	mockIndex.Ellip("%d", 0)
	mockIndex.Ellip("%d")
	mockIndex.Ellip("%d")
}

func TestGrabPointer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIndex := NewMockIndex(ctrl)
	mockIndex.EXPECT().Ptr(gomock.Any()).SetArg(0, 7) // set first argument to 7

	i := user.GrabPointer(mockIndex)
	if i != 7 {
		t.Errorf("Expected 7, got %d", i)
	}
}

func TestEmbeddedInterface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEmbed := NewMockEmbed(ctrl)
	mockEmbed.EXPECT().RegularMethod()
	mockEmbed.EXPECT().EmbeddedMethod()
	mockEmbed.EXPECT().ForeignEmbeddedMethod()

	mockEmbed.RegularMethod()
	mockEmbed.EmbeddedMethod()
	var emb imp1.ForeignEmbedded = mockEmbed // also does interface check
	emb.ForeignEmbeddedMethod()
}

func TestExpectTrueNil(t *testing.T) {
	// Make sure that passing "nil" to EXPECT (thus as a nil interface value),
	// will correctly match a nil concrete type.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIndex := NewMockIndex(ctrl)
	mockIndex.EXPECT().Ptr(nil) // this nil is a nil any
	mockIndex.Ptr(nil)          // this nil is a nil *int
}

func TestDoAndReturnSignature(t *testing.T) {
	t.Run("wrong number of return args", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockIndex := NewMockIndex(ctrl)

		mockIndex.EXPECT().Slice(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ []int, _ []byte) {},
		)

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic")
			}
		}()

		mockIndex.Slice([]int{0}, []byte("meow"))
	})

	t.Run("wrong type of return arg", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockIndex := NewMockIndex(ctrl)

		mockIndex.EXPECT().Slice(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ []int, _ []byte) bool {
				return true
			})

		mockIndex.Slice([]int{0}, []byte("meow"))
	})
}

func TestExpectCondForeignFour(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIndex := NewMockIndex(ctrl)
	mockIndex.EXPECT().ForeignFour(gomock.Cond(func(x any) bool {
		four, ok := x.(imp_four.Imp4)
		if !ok {
			return false
		}
		return four.Field == "Cool"
	}))

	mockIndex.ForeignFour(imp_four.Imp4{Field: "Cool"})
}
