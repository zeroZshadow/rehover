#include "Game.h"
#include "behaviours/Hovercraft.h"
#include "components/Camera.h"
#include "components/Renderable.h"
#include "components/Transform.h"
#include "input/HovercraftController.h"
#include "systems/RenderSystem.h"

namespace cp = Components;
namespace bh = Behaviours;

Game::Game() {
	input = systems.add<InputSystem>();
	systems.add<bh::HovercraftSystem>();
	systems.add<RenderSystem>();
	systems.configure();
}

void Game::init(Mesh* mesh) {
	// Controller (for hovercraft)
	auto controller = std::make_shared<GCHovercraftController>(input->GetController(0));
	// Hovercraft
	hovercraft = entities.create();
	hovercraft.assign<cp::Transform>(cp::Transform({0, 0, 0}));
	hovercraft.assign<cp::Renderable>(cp::Renderable(mesh));
	hovercraft.assign<bh::Hovercraft>(bh::Hovercraft{controller});

	// Camera
	ex::Entity camera = entities.create();
	camera.assign<cp::Transform>(cp::Transform({0, 0, -10}));
	camera.assign<cp::Camera>(cp::Camera());
}

void Game::update(ex::TimeDelta dt) { systems.update_all(dt); }