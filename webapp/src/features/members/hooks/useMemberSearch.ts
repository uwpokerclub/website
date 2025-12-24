import { useState, useCallback } from "react";
import { searchMembers as searchMembersApi } from "../api/memberRegistrationApi";
import { User } from "../../../types/user";

// Member type alias for semantic clarity (API returns User type)
type Member = User;

export type MemberSearchOption = {
  value: string;
  label: string;
  data: Member;
};

type UseMemberSearchReturn = {
  searchMembers: (query: string) => Promise<MemberSearchOption[]>;
  isSearching: boolean;
  error: string | null;
  clearError: () => void;
};

/**
 * Hook for searching members with state management
 * Designed to work with Combobox component which handles debouncing
 */
export function useMemberSearch(): UseMemberSearchReturn {
  const [isSearching, setIsSearching] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const searchMembers = useCallback(async (query: string): Promise<MemberSearchOption[]> => {
    setError(null);
    setIsSearching(true);

    try {
      const result = await searchMembersApi(query);

      if (!result.success) {
        setError(result.error);
        return [];
      }

      // Transform Member[] to ComboboxOption format
      // Note: API returns id as number, convert to string for form compatibility
      return result.data.map((member) => ({
        value: String(member.id),
        label: `${member.firstName} ${member.lastName} (${member.email})`,
        data: member,
      }));
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "An unexpected error occurred";
      setError(errorMessage);
      return [];
    } finally {
      setIsSearching(false);
    }
  }, []);

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  return {
    searchMembers,
    isSearching,
    error,
    clearError,
  };
}
