import { useState, useContext, useCallback, useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Modal, Button, FormField, Input, Select, Textarea, Spinner, useToast } from "@uwpokerclub/components";
import { SemesterContext } from "../../../../contexts";
import { editEventSchema, POKER_FORMATS, type EditEventFormData } from "../../schemas/eventSchema";
import { fetchEvent, updateEvent } from "../../api/eventApi";
import styles from "./EditEventModal.module.css";

// Transform POKER_FORMATS to Select options
const formatOptions = POKER_FORMATS.map((format) => ({
  value: format,
  label: format,
}));

/**
 * Basic event data from the table row
 */
export interface EventData {
  id: number;
  name: string;
  format: string;
  notes: string;
  startDate: string;
  state: number;
}

export interface EditEventModalProps {
  isOpen: boolean;
  event: EventData | null;
  onClose: () => void;
  onSuccess: () => void;
}

/**
 * EditEventModal - Modal for editing existing events
 *
 * Fetches full event details on open and pre-populates the form.
 * Uses SemesterContext for API calls.
 */
export function EditEventModal({ isOpen, event, onClose, onSuccess }: EditEventModalProps) {
  const semesterContext = useContext(SemesterContext);
  const { showToast } = useToast();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isLoadingEvent, setIsLoadingEvent] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);

  const form = useForm<EditEventFormData>({
    resolver: zodResolver(editEventSchema),
    defaultValues: {
      name: "",
      startDate: "",
      format: "" as EditEventFormData["format"],
      pointsMultiplier: 1,
      notes: "",
    },
  });

  // Fetch full event details and pre-populate form when modal opens
  useEffect(() => {
    if (!isOpen || !event || !semesterContext?.currentSemester?.id) {
      return;
    }

    let mounted = true;

    const loadEventDetails = async () => {
      setIsLoadingEvent(true);
      setSubmitError(null);

      const result = await fetchEvent(semesterContext.currentSemester!.id, event.id);

      if (mounted) {
        if (result.success) {
          const eventData = result.data;
          // Format the date for datetime-local input
          const startDate = new Date(eventData.startDate);
          const formattedDate = startDate.toISOString().slice(0, 16);

          form.reset({
            name: eventData.name,
            startDate: formattedDate,
            format: eventData.format as EditEventFormData["format"],
            pointsMultiplier: eventData.pointsMultiplier,
            notes: eventData.notes || "",
          });
        } else {
          setSubmitError(`Failed to load event details: ${result.error}`);
        }
        setIsLoadingEvent(false);
      }
    };

    loadEventDetails();

    return () => {
      mounted = false;
    };
  }, [isOpen, event, semesterContext?.currentSemester, form]);

  // Handle modal close
  const handleClose = useCallback(() => {
    form.reset();
    setSubmitError(null);
    onClose();
  }, [form, onClose]);

  // Handle form submit
  const handleSubmit = async (data: EditEventFormData) => {
    if (!semesterContext?.currentSemester?.id || !event) {
      setSubmitError("No semester or event selected");
      return;
    }

    setIsSubmitting(true);
    setSubmitError(null);

    const result = await updateEvent(semesterContext.currentSemester.id, event.id, {
      name: data.name,
      format: data.format,
      notes: data.notes || "",
      startDate: new Date(data.startDate).toISOString(),
      pointsMultiplier: data.pointsMultiplier,
    });

    setIsSubmitting(false);

    if (result.success) {
      showToast({
        message: `"${data.name}" updated successfully!`,
        variant: "success",
        duration: 3000,
      });
      onSuccess();
      handleClose();
    } else {
      setSubmitError(result.error);
      showToast({
        message: result.error,
        variant: "error",
        duration: 5000,
      });
    }
  };

  // Footer with actions
  const footer = (
    <div className={styles.footer}>
      <Button variant="tertiary" onClick={handleClose} disabled={isSubmitting || isLoadingEvent}>
        Cancel
      </Button>
      <Button type="submit" form="edit-event-form" disabled={isSubmitting || isLoadingEvent}>
        {isSubmitting ? "Saving..." : "Save Changes"}
      </Button>
    </div>
  );

  const {
    register,
    formState: { errors },
  } = form;

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title="Edit Event" size="lg" footer={footer}>
      <div className={styles.content}>
        {/* Error display */}
        {submitError && <div className={styles.errorAlert}>{submitError}</div>}

        {isLoadingEvent ? (
          <div className={styles.loadingState}>
            <Spinner size="lg" />
            <p>Loading event details...</p>
          </div>
        ) : (
          <form id="edit-event-form" onSubmit={form.handleSubmit(handleSubmit)} noValidate>
            <div className={styles.form}>
              <div className={styles.fullWidth}>
                <FormField label="Event Name" htmlFor="name" required error={errors.name?.message}>
                  {(props) => (
                    <Input
                      {...props}
                      {...register("name")}
                      type="text"
                      placeholder="e.g., Tournament #5"
                      error={!!errors.name}
                      fullWidth
                    />
                  )}
                </FormField>
              </div>

              <FormField label="Start Date" htmlFor="startDate" required error={errors.startDate?.message}>
                {(props) => (
                  <Input
                    {...props}
                    {...register("startDate")}
                    type="datetime-local"
                    error={!!errors.startDate}
                    fullWidth
                  />
                )}
              </FormField>

              <FormField label="Format" htmlFor="format" required error={errors.format?.message}>
                {(props) => (
                  <Select
                    {...props}
                    {...register("format")}
                    options={formatOptions}
                    placeholder="Select a format"
                    error={!!errors.format}
                    fullWidth
                  />
                )}
              </FormField>

              <FormField
                label="Points Multiplier"
                htmlFor="pointsMultiplier"
                required
                error={errors.pointsMultiplier?.message}
              >
                {(props) => (
                  <Input
                    {...props}
                    {...register("pointsMultiplier", { valueAsNumber: true })}
                    type="number"
                    min={0}
                    step={0.5}
                    placeholder="1"
                    error={!!errors.pointsMultiplier}
                    fullWidth
                  />
                )}
              </FormField>

              <div className={styles.fullWidth}>
                <FormField label="Additional Details" htmlFor="notes">
                  {(props) => (
                    <Textarea
                      {...props}
                      {...register("notes")}
                      placeholder="Optional notes about the event..."
                      rows={3}
                      fullWidth
                    />
                  )}
                </FormField>
              </div>
            </div>
          </form>
        )}
      </div>
    </Modal>
  );
}
