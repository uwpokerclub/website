import { useQuery, useMutation, useQueryClient, keepPreviousData } from "@tanstack/react-query";
import {
  fetchMemberships,
  createMember,
  createMembership,
  registerNewMemberWithMembership,
  updateMember,
  updateMembership,
  deleteMember,
  deleteMembership,
  type FetchMembershipsParams,
  type UpdateMemberRequest,
  type UpdateMembershipRequest,
} from "../api/memberRegistrationApi";
import type { CreateMemberFormData } from "../validation/registrationSchema";

export const memberKeys = {
  all: ["members"] as const,
  bySemester: (semesterId: string) => [...memberKeys.all, semesterId] as const,
  list: (semesterId: string, params: FetchMembershipsParams) => [...memberKeys.bySemester(semesterId), params] as const,
};

export function useMemberships(semesterId: string | undefined, params: FetchMembershipsParams) {
  return useQuery({
    queryKey: memberKeys.list(semesterId ?? "", params),
    queryFn: () => fetchMemberships(semesterId!, params),
    enabled: !!semesterId,
    placeholderData: keepPreviousData,
  });
}

export function useCreateMember() {
  // No invalidation: a bare member without a membership doesn't appear in any
  // list; useCreateMembership / useRegisterNewMemberWithMembership handle that.
  return useMutation({
    mutationFn: (data: CreateMemberFormData) => createMember(data),
  });
}

export function useCreateMembership() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      semesterId,
      memberId,
      paid,
      discounted,
    }: {
      semesterId: string;
      memberId: string;
      paid: boolean;
      discounted: boolean;
    }) => createMembership(semesterId, memberId, paid, discounted),
    onSuccess: (_data, { semesterId }) => {
      queryClient.invalidateQueries({ queryKey: memberKeys.bySemester(semesterId) });
    },
  });
}

export function useRegisterNewMemberWithMembership() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      memberData,
      semesterId,
      paid,
      discounted,
    }: {
      memberData: CreateMemberFormData;
      semesterId: string;
      paid: boolean;
      discounted: boolean;
    }) => registerNewMemberWithMembership(memberData, semesterId, paid, discounted),
    onSuccess: (_data, { semesterId }) => {
      queryClient.invalidateQueries({ queryKey: memberKeys.bySemester(semesterId) });
    },
  });
}

export function useUpdateMember() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ memberId, data }: { memberId: string; data: UpdateMemberRequest; semesterId?: string }) =>
      updateMember(memberId, data),
    onSuccess: (_data, { semesterId }) => {
      queryClient.invalidateQueries({
        queryKey: semesterId ? memberKeys.bySemester(semesterId) : memberKeys.all,
      });
    },
  });
}

export function useUpdateMembership() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      semesterId,
      membershipId,
      data,
    }: {
      semesterId: string;
      membershipId: string;
      data: UpdateMembershipRequest;
    }) => updateMembership(semesterId, membershipId, data),
    onSuccess: (_data, { semesterId }) => {
      queryClient.invalidateQueries({ queryKey: memberKeys.bySemester(semesterId) });
    },
  });
}

export function useDeleteMember() {
  const queryClient = useQueryClient();
  // Intentional broad invalidation: a deleted user disappears from every
  // semester's membership list, so all member-keyed queries must refetch.
  return useMutation({
    mutationFn: (memberId: string) => deleteMember(memberId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: memberKeys.all });
    },
  });
}

export function useDeleteMembership() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ semesterId, membershipId }: { semesterId: string; membershipId: string }) =>
      deleteMembership(semesterId, membershipId),
    onSuccess: (_data, { semesterId }) => {
      queryClient.invalidateQueries({ queryKey: memberKeys.bySemester(semesterId) });
    },
  });
}
