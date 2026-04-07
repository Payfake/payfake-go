package payfake

// pageOrDefault returns the page number or 1 if not set.
func pageOrDefault(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

// perPageOrDefault returns the per page count or 50 if not set.
func perPageOrDefault(perPage int) int {
	if perPage < 1 {
		return 50
	}
	if perPage > 100 {
		return 100
	}
	return perPage
}
