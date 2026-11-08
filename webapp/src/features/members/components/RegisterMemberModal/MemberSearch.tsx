import { useState, useCallback } from "react";
import { Combobox, type ComboboxOption } from "@uwpokerclub/components";
import { useMemberSearch, type MemberSearchOption } from "../../hooks/useMemberSearch";

/**
 * Selected member data passed to onSelect callback
 */
export interface SelectedMemberData {
  id: string;
  firstName: string;
  lastName: string;
  email?: string;
  faculty?: string;
  questId?: string;
}

export interface MemberSearchProps {
  /**
   * Callback when a member is selected. Receives full member data or null if cleared.
   */
  onSelect: (member: SelectedMemberData | null) => void;
  /**
   * Currently selected member ID
   */
  value: string | null;
  /**
   * External error message (e.g., from form validation)
   */
  error?: string;
  includeIdentityDetails?: boolean;
}

/**
 * MemberSearch component - Searchable dropdown for finding existing members
 *
 * Uses Combobox from component library with async search support.
 * Searches by name, email, or student ID.
 */
export function MemberSearch({
  onSelect,
  value,
  error: externalError,
  includeIdentityDetails = false,
}: MemberSearchProps) {
  const { searchMembers, isSearching, error: searchError, clearError } = useMemberSearch();
  const [options, setOptions] = useState<ComboboxOption[]>([]);
  const [memberDataMap, setMemberDataMap] = useState<Map<string, SelectedMemberData>>(new Map());

  const handleSearch = useCallback(
    async (query: string) => {
      clearError();
      const results = await searchMembers(query);
      // Transform MemberSearchOption[] to ComboboxOption[]
      setOptions(
        results.map((result: MemberSearchOption) => ({
          value: result.value,
          label: includeIdentityDetails
            ? `${result.data.firstName} ${result.data.lastName} (${result.data.questId || "Quest ID missing"} · ${result.data.faculty || "Faculty unavailable"})`
            : result.label,
        })),
      );
      // Store member data for lookup when selected
      const dataMap = new Map<string, SelectedMemberData>();
      results.forEach((result: MemberSearchOption) => {
        dataMap.set(result.value, {
          id: result.value,
          firstName: result.data.firstName,
          lastName: result.data.lastName,
          email: result.data.email,
          faculty: result.data.faculty,
          questId: result.data.questId,
        });
      });
      setMemberDataMap(dataMap);
    },
    [searchMembers, clearError, includeIdentityDetails],
  );

  const handleChange = useCallback(
    (selectedValue: string | null) => {
      if (selectedValue === null) {
        onSelect(null);
      } else {
        const memberData = memberDataMap.get(selectedValue);
        onSelect(memberData ?? { id: selectedValue, firstName: "", lastName: "" });
      }
    },
    [onSelect, memberDataMap],
  );

  // Combine external and search errors
  const displayError = externalError || searchError;

  return (
    <Combobox
      data-qa="member-search-combobox"
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
