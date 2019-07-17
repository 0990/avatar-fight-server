package game

import "math"

func NewRotation(lastRotation, pressTime, targetRotation float32) float32 {
	delta := math.Abs(float64(targetRotation - lastRotation))
	if delta >= 180 {
		delta = math.Abs(360 - delta)
	}
	var newRotation float32
	if delta <= float64(pressTime*ROTATION_DELTA) {
		newRotation = targetRotation
	} else {
		var change float32
		if lastRotation >= 0 {
			if lastRotation-180 < targetRotation && targetRotation < lastRotation {
				change = -pressTime * ROTATION_DELTA
			} else {
				change = pressTime * ROTATION_DELTA
			}
		} else {
			if lastRotation < targetRotation && targetRotation < lastRotation+180 {
				change = pressTime * ROTATION_DELTA
			} else {
				change = -pressTime * ROTATION_DELTA
			}
		}
		newRotation = lastRotation + change
	}

	if newRotation > 180 {
		newRotation = newRotation - 360
	}
	if newRotation < -180 {
		newRotation = 360 + newRotation
	}
	return newRotation
}

func NewXPos(lastPos, lastRotation, pressTime float32) float32 {
	newXPos := lastPos + pressTime*ENTITY_SPEED*float32(math.Cos(float64(lastRotation*math.Pi/180)))
	if newXPos < ENTITY_RADIUS {
		newXPos = ENTITY_RADIUS
	}
	if newXPos > WORLD_WIDTH-ENTITY_RADIUS {
		newXPos = WORLD_WIDTH - ENTITY_RADIUS
	}
	return newXPos
}

func NewYPos(lastPos, lastRotation, pressTime float32) float32 {
	newYPos := lastPos + pressTime*ENTITY_SPEED*float32(math.Sin(float64(lastRotation*math.Pi/180)))
	if newYPos < ENTITY_RADIUS {
		newYPos = ENTITY_RADIUS
	}

	if newYPos > WORLD_HEIGHT-ENTITY_RADIUS {
		newYPos = WORLD_HEIGHT - ENTITY_RADIUS
	}
	return newYPos
}
