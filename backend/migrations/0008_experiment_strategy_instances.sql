-- Replaces the flat strategies[]+params{} pair (one shared params bag for
-- every strategy in a candidate) with one instances[] array, each entry
-- carrying its own type/params/weight — lets a composite combine multiple
-- configured instances (including repeats of the same type, e.g. MA(20)
-- and MA(50) together) instead of collapsing them into one shared bag.
-- Existing rows lose their old strategies/params detail (acceptable: this
-- is dev/demo data, not graded historical state) but keep every other
-- column untouched.
ALTER TABLE experiments ADD COLUMN instances TEXT NOT NULL DEFAULT '[]';
ALTER TABLE experiments DROP COLUMN strategies;
ALTER TABLE experiments DROP COLUMN params;
