For example:
func signature:
/// Add three numbers together.
/// This function takes three numbers as input and returns the sum of the three numbers.
func Add3Numbers(x int, y int, z int) int {

unit tests:
func TestAdd(t *testing.T) {
    assert := assert.New(t)
    assert.Equal(7, Add3Numbers(2, 3+rand.Intn(1000)*0, 2))
    assert.Equal(15, Add3Numbers(5, 7, 3))
}
