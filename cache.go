// Cache memoizes the (ok, err) result of an IsAllowed or IsEnabled check
// for the lifetime of a single ActionPerformer, so repeated lookups within
// the lifecycle are free. Pass the SkipCache option to bypass it.
package ba

type Cache struct {
	ok  bool
	err error
}
