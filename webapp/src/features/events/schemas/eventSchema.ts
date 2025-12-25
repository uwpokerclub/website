import { z } from "zod";

// Poker format options (matching backend expectations)
export const POKER_FORMATS = [
  "No Limit Hold'em",
  "Pot Limit Omaha",
  "Short Deck No Limit Hold'em",
  "Dealers Choice",
] as const;

export const formatSchema = z.enum(POKER_FORMATS, {
  errorMap: () => ({ message: "Please select a format" }),
});

export type PokerFormat = z.infer<typeof formatSchema>;

// Single blind level schema
export const blindLevelSchema = z.object({
  small: z.number().min(0, "Small blind must be >= 0"),
  big: z.number().min(0, "Big blind must be >= 0"),
  ante: z.number().min(0, "Ante must be >= 0"),
  time: z.number().min(1, "Time must be at least 1 minute"),
});

export type BlindLevel = z.infer<typeof blindLevelSchema>;

// Schema for selecting existing structure
export const selectStructureSchema = z.object({
  mode: z.literal("select"),
  structureId: z.number().min(1, "Please select a structure"),
});

// Schema for creating new structure
export const createStructureSchema = z.object({
  mode: z.literal("create"),
  name: z.string().min(1, "Structure name is required"),
  blinds: z.array(blindLevelSchema).min(1, "At least one blind level is required"),
});

// Union type for structure configuration
export const structureConfigSchema = z.discriminatedUnion("mode", [selectStructureSchema, createStructureSchema]);

export type StructureConfig = z.infer<typeof structureConfigSchema>;

// Main event creation schema
export const createEventSchema = z.object({
  name: z.string().min(1, "Event name is required"),
  startDate: z.string().min(1, "Start date is required"),
  format: formatSchema,
  pointsMultiplier: z.number().min(0, "Points multiplier must be >= 0"),
  notes: z.string().optional(),
  structure: structureConfigSchema,
});

export type CreateEventFormData = z.infer<typeof createEventSchema>;
