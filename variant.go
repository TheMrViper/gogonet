package gogonet

const (
	NIL = iota

	// atomic types
	BOOL
	INT
	REAL
	STRING

	// math types

	VECTOR2 // 5
	RECT2
	VECTOR3
	TRANSFORM2D
	PLANE
	QUAT // 10
	AABB
	BASIS
	TRANSFORM

	// misc types
	COLOR
	NODE_PATH // 15
	_RID
	OBJECT
	DICTIONARY
	ARRAY

	// arrays
	POOL_BYTE_ARRAY // 20
	POOL_INT_ARRAY
	POOL_REAL_ARRAY
	POOL_STRING_ARRAY
	POOL_VECTOR2_ARRAY
	POOL_VECTOR3_ARRAY // 25
	POOL_COLOR_ARRAY

	VARIANT_MAX
)
