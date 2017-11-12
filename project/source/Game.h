#pragma once
#include "rendering/Mesh.h"
#include "systems/InputSystem.h"
#include <entityx/entityx.h>
#include <memory>

namespace ex = entityx;

class Game : public ex::EntityX {
public:
	explicit Game();
	void init(Mesh* mesh);
	void update(ex::TimeDelta dt);

	ex::Entity hovercraft;
	std::shared_ptr<InputSystem> input;
};