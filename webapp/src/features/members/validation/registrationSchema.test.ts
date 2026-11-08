import {
  createMemberSchema,
  editMemberSchema,
  executiveAddMemberSchema,
  isValidQuestId,
  normalizeQuestIdInput,
} from "./registrationSchema";

const createMember = {
  id: "12345678",
  firstName: "Taylor",
  lastName: "Member",
  email: "taylor@example.com",
  faculty: "Math",
};

describe("Quest ID validation", () => {
  it("validates normalized Quest ID input on create", () => {
    const result = createMemberSchema.safeParse({ ...createMember, questId: "  AbC123  " });

    expect(result.success).toBe(true);
    expect(normalizeQuestIdInput("  AbC123  ")).toBe("abc123");
  });

  it.each(["bad.id", "bad-id", "bad id", "bad_id"])('rejects unsupported Quest ID "%s"', (questId) => {
    expect(createMemberSchema.safeParse({ ...createMember, questId }).success).toBe(false);
  });

  it("accepts omitted, empty, and whitespace-only optional Quest IDs", () => {
    expect(createMemberSchema.safeParse(createMember).success).toBe(true);
    expect(createMemberSchema.safeParse({ ...createMember, questId: "" }).success).toBe(true);
    expect(createMemberSchema.safeParse({ ...createMember, questId: "   " }).success).toBe(true);
  });

  it("requires and normalizes a valid Quest ID for executive additions only", () => {
    expect(executiveAddMemberSchema.safeParse(createMember).success).toBe(false);
    expect(executiveAddMemberSchema.safeParse({ ...createMember, questId: " NewID7 " }).success).toBe(true);
    expect(createMemberSchema.safeParse(createMember).success).toBe(true);
  });

  it("allows legacy values through edit form validation while checking changed input separately", () => {
    expect(
      editMemberSchema.safeParse({
        questId: " Legacy.ID ",
        firstName: "Taylor",
        lastName: "Member",
        email: "taylor@example.com",
        faculty: "Math",
      }).success,
    ).toBe(true);
    expect(normalizeQuestIdInput(" Legacy.ID ")).toBe("legacy.id");
    expect(isValidQuestId(normalizeQuestIdInput(" Legacy.ID "))).toBe(false);
    expect(isValidQuestId(normalizeQuestIdInput(" NewID7 "))).toBe(true);
  });
});
