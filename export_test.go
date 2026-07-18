package shogi

//this test file is export private method

// Package
var ExportNewPos = newPos
var ExportCanPawn = canPawn
var ExportCanKnight = canKnight
var ExportCanSilver = canSilver
var ExportCanGold = canGold
var ExportCanKing = canKing

// attack.go
var ExportAttacks = attacks
var ExportHasMoveMask = hasMoveMask
var ExportSquareOf = squareOf

// Board
var ExportBoardSet = (*Board).set
var ExportBoardParse = (*Board).parse
var ExportBoardPseudoCandidate = (*Board).pseudoCandidate

// CampBoard
var ExportCampBoardCanBit = (*CampBoard).canBit
var ExportCampBoardSet = (*CampBoard).set
var ExportCampBoardCopy = (*CampBoard).copy
var ExportCampBoardSetEnemy = (*CampBoard).setEnemy

// BitBoard
var ExportBitBoardGet = (*BitBoard).get
var ExportBitBoardIs = (*BitBoard).is
var ExportBitBoardSet = (*BitBoard).set
var ExportBitBoardClear = (*BitBoard).clear
