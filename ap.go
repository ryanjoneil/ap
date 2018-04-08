package ap

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */
/*                                                                           */
/*  This file is part of the ap.go library for incremental solution of       */
/*  assignments problems. The code is a Go port of the original Fortran      */
/*  from the following paper:                                                */
/*                                                                           */
/*      Giorgio Carpaneto, Silvano Martello, and Paolo Toth.                 */
/*      "Algorithms and codes for the assignment problem."                   */
/*      Annals of Operations Research 13, no. 1 (1988): 191-223.             */
/*                                                                           */
/*  ap.go is distributed under the terms of the MIT License. You should      */
/*  have received a copy with the source If not, email roneil1@gmu.edu.      */
/*                                                                           */
/*      Copyright (C) 2018 Ryan J. O'Neil <roneil1@gmu.edu>                  */
/*                                                                           */
/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// AP represents a Primal-Dual Assignment Problem.
type AP struct {
	A [][]int64 // A[i][j] = cost of assigning row i to column j
	U []int64   // U[i] = dual price for row i
	V []int64   // V[j] = dual price for column j
	Z int64     // Objective value

	f    []int   // f[i] = column assigned to row i, -1 if unassigned
	fBar []int   // fBar[j] = row assigned to column j, -1 if unassigned
	p    []int   // First column of row i not yet examined in phase 2
	c    []int   // c[j] = row preceding column j in current alternating path
	pi   []int64 // pi[j] = min { a[i,j] - u[i] - v[j] | i in lr, i != fBar[j] }

	Size        int
	initialized bool // true if row and column reductions have been done
}

// X returns true if row i is assigned to column j, false if not.
func (ap *AP) X(i, j int) bool {
	return ap.Col(i) == j
}

// Col returns the column assigned to row i, -1 if none assigned.
func (ap *AP) Col(i int) int {
	return ap.f[i]
}

// Row returns the row assigned to column j, -1 if none assigned.
func (ap *AP) Row(i int) int {
	return ap.fBar[i]
}

// RC returns the reduced cost of an arc.
func (ap *AP) RC(i, j int) int64 {
	return ap.A[i][j] - ap.U[i] - ap.V[j]
}

// Remove takes an arc out of the solution and gives it a new objective value.
func (ap *AP) Remove(i, j int, obj int64) {
	if ap.f[i] == j {
		ap.f[i] = -1
		ap.fBar[j] = -1
	}
	ap.A[i][j] = obj
}
