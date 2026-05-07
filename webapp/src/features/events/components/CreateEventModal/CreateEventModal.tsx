import { useState, useContext, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { useForm, FormProvider } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Modal, Button, useToast } from "@uwpokerclub/components";
import { SemesterContext } from "../../../../contexts";
import { EventDetailsForm } from "./EventDetailsForm";
import { StructureSelector } from "./StructureSelector";
import { createEventSchema, type CreateEventFormData } from "../../schemas/eventSchema";
import { useCreateEvent } from "../../hooks/useEventQueries";
import { useStructures, useCreateStructure } from "@/features/structures/hooks/useStructureQueries";
import styles from "./CreateEventModal.module.css";

export interface CreateEventModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess?: () => void;
}

/**
 * CreateEventModal - Modal for creating new events
 *
 * Supports selecting existing structure or creating a new one inline.
 * Uses SemesterContext for automatic semester assignment.
 */
export function CreateEventModal({ isOpen, onClose, onSuccess }: CreateEventModalProps) {
  const navigate = useNavigate();
  const semesterContext = useContext(SemesterContext);
  const { showToast } = useToast();

  // Only fetch structures when the modal is open
  const { data: structures = [], isLoading: isLoadingStructures, error: structuresError } = useStructures();
  const createStructureMutation = useCreateStructure();
  const createEventMutation = useCreateEvent();

  const structureFetchError = isOpen && structuresError ? (structuresError.message ?? null) : null;
  const isSubmitting = createStructureMutation.isPending || createEventMutation.isPending;
  const [submitError, setSubmitError] = useState<string | null>(null);

  const form = useForm<CreateEventFormData>({
    resolver: zodResolver(createEventSchema),
    defaultValues: {
      name: "",
      startDate: "",
      format: "" as CreateEventFormData["format"],
      pointsMultiplier: 1,
      notes: "",
      structure: {
        mode: "select",
        structureId: 0,
      },
    },
  });

  // Reset form when modal closes
  const handleClose = useCallback(() => {
    form.reset();
    setSubmitError(null);
    onClose();
  }, [form, onClose]);

  // Handle form submission
  const handleSubmit = async (data: CreateEventFormData) => {
    if (!semesterContext?.currentSemester?.id) {
      setSubmitError("No semester selected. Please select a semester first.");
      return;
    }

    setSubmitError(null);

    try {
      let structureId: number;

      // If creating new structure, create it first
      if (data.structure.mode === "create") {
        const structure = await createStructureMutation.mutateAsync({
          name: data.structure.name,
          blinds: data.structure.blinds,
        });
        structureId = structure.id;
      } else {
        structureId = data.structure.structureId;
      }

      // Create the event
      const createdEvent = await createEventMutation.mutateAsync({
        semesterId: semesterContext.currentSemester.id,
        data: {
          name: data.name,
          format: data.format,
          notes: data.notes || "",
          startDate: new Date(data.startDate),
          structureId,
          pointsMultiplier: data.pointsMultiplier,
        },
      });

      // Success!
      showToast({
        message: `Event "${data.name}" created successfully!`,
        variant: "success",
        duration: 3000,
      });

      onSuccess?.();
      handleClose();

      // Navigate to the new event page
      navigate(`/admin/events/${createdEvent.id}`);
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : "An unexpected error occurred. Please try again.");
    }
  };

  // Footer with actions
  const footer = (
    <div className={styles.footer}>
      <Button variant="tertiary" onClick={handleClose} disabled={isSubmitting} data-qa="create-event-cancel-btn">
        Cancel
      </Button>
      <Button type="submit" form="create-event-form" disabled={isSubmitting} data-qa="create-event-submit-btn">
        {isSubmitting ? "Creating..." : "Create Event"}
      </Button>
    </div>
  );

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="Create Event"
      size="lg"
      footer={footer}
      data-qa="create-event-modal"
    >
      <div className={styles.content}>
        {/* Error display */}
        {submitError && (
          <div className={styles.errorAlert} data-qa="create-event-error-alert">
            {submitError}
          </div>
        )}
        {structureFetchError && (
          <div className={styles.errorAlert} data-qa="create-event-structure-error-alert">
            Failed to load structures: {structureFetchError}
          </div>
        )}

        <FormProvider {...form}>
          <form id="create-event-form" onSubmit={form.handleSubmit(handleSubmit)} noValidate>
            <div className={styles.section}>
              <EventDetailsForm />
            </div>

            <StructureSelector structures={structures} isLoadingStructures={isLoadingStructures} />
          </form>
        </FormProvider>
      </div>
    </Modal>
  );
}
