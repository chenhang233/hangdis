package database

const (
	upsertPolicy  = iota // default
	insertPolicy         // set nx Only add new elements Don't update already existing elements.
	updatePolicy         // set xx Only update elements that already exist. Don't add new elements.
	greaterExpiry        // set GL
	lessExpiry           // set LT
)
const unlimitedTTL int64 = 0
