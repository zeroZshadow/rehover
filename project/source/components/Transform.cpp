#include "Transform.h"
#include "../utils/math.h"

namespace Components {
Mtx& Transform::GetMatrix() {
	c_guMtxQuat(matrix, &rotation);
	guMtxScaleApply(matrix, matrix, scale.x, scale.y, scale.z);
	guMtxTransApply(matrix, matrix, position.x, position.y, position.z);

	guMtxIdentity(matrix);
	guMtxTrans(matrix, position.x, position.y, position.z);
	return matrix;
}

void Transform::SetRotation(guVector rotation) { this->rotation = Math::EulerToQuaternion(rotation); }

void Transform::SetRotation(guQuaternion rotation) { this->rotation = rotation; }

void Transform::Lookat(guVector target) {
	Mtx temp;
	guVector up = {0, 1, 0};
	guLookAt(temp, &position, &up, &target);
	c_guQuatMtx(&rotation, temp);
}
}; // namespace Components