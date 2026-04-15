-- +migrate Up

-- Add bdd_scenarios to the allowed phases for refinement sessions and messages.

ALTER TABLE refinement_sessions
    DROP CONSTRAINT refinement_sessions_current_phase_check;
ALTER TABLE refinement_sessions
    ADD CONSTRAINT refinement_sessions_current_phase_check
    CHECK (current_phase IN ('actor_goal', 'main_scenario', 'extensions', 'acceptance_criteria', 'bdd_scenarios'));

ALTER TABLE refinement_messages
    DROP CONSTRAINT refinement_messages_phase_check;
ALTER TABLE refinement_messages
    ADD CONSTRAINT refinement_messages_phase_check
    CHECK (phase IN ('actor_goal', 'main_scenario', 'extensions', 'acceptance_criteria', 'bdd_scenarios'));
