import { useState, useCallback } from "react";
import { Combobox, type ComboboxOption } from "@uwpokerclub/components";
import { useMemberSearch, type MemberSearchOption } from "../../hooks/useMemberSearch";

export interface MemberSearchProps {
  /**
   * Callback when a member is selected
   */
  onSelect: (memberId: string | null) => void;
  /**
   * Currently selected member ID
   */
  value: string | null;
  /**
   * External error message (e.g., from form validation)
   */
  error?: string;
}

/**
 * MemberSearch component - Searchable dropdown for finding existing members
 *
 * Uses Combobox from component library with async search support.
 * Searches by name, email, or student ID.
 */
export function MemberSearch({ onSelect, value, error: externalError }: MemberSearchProps) {
  const { searchMembers, isSearching, error: searchError, clearError } = useMemberSearch();
  const [options, setOptions] = useState<ComboboxOption[]>([]);

  const handleSearch = useCallback(
    async (query: string) => {
      clearError();
      const results = await searchMembers(query);
      // Transform MemberSearchOption[] to ComboboxOption[]
      setOptions(
        results.map((result: MemberSearchOption) => ({
          value: result.value,
          label: result.label,
        })),
      );
    },
    [searchMembers, clearError],
  );

  const handleChange = useCallback(
    (selectedValue: string | null) => {
      onSelect(selectedValue);
    },
    [onSelect],
  );

  // Combine external and search errors
  const displayError = externalError || searchError;

  return (
    <Combobox
      options={options}
      onSearch={handleSearch}
      onChange={handleChange}
      value={value}
      placeholder="Search by name, email, or student ID..."
      isLoading={isSearching}
      debounceMs={300}
      emptyMessage="No members found"
      error={!!displayError}
      errorMessage={displayError || undefined}
      fullWidth
    />
  );
}
