// Package dice parses and rolls tabletop RPG dice expressions.
//
// Supported notation:
//
//	XdY                two arguments form: roll X dice with Y sides
//	dY                 implicit count of 1
//	XdY+Z, XdY-Z       arithmetic with integer modifiers
//	XdYkhN, XdYklN     keep highest / lowest N (default 1)
//	XdYdhN, XdYdlN     drop highest / lowest N (default 1)
//	XdYx[comp], xo     exploding dice (recursive / once)
//	XdYr<comp>, rr     reroll once / recursively
//	XdYcs<comp>        count successes
//	XdYcf<comp>        count failures (subtracts)
//	adv(expr)          roll twice, take the higher
//	dis(expr)          roll twice, take the lower
//	(expr)             grouping
//
// Comparisons take the form >=N, <=N, >N, <N, =N, or bare N (interpreted
// per modifier: explode defaults to ">=N or =max", reroll requires explicit).
//
// The default randomness source is crypto/rand. Tests and applications that
// need determinism can supply their own Roller via RollWith.
package dice
