import { useFormContext, useWatch } from "react-hook-form";
import { FormField, Input, Select, Spinner, Radio } from "@uwpokerclub/components";
import { useAuth } from "../../../../hooks";
import type { Structure } from "../../../../types";
import type { CreateEventFormData } from "../../schemas/eventSchema";
import { BlindLevelsTable } from "./BlindLevelsTable";
import styles from "./StructureSelector.module.css";

interface StructureSelectorProps {
  structures: Structure[];
  isLoadingStructures: boolean;
}

/**
 * StructureSelector - Radio-based selector for choosing existing or creating new structure
 *
 * Permission-gated: "Create new structure" option only shown if user has create.structure permission
 */
export function StructureSelector({ structures, isLoadingStructures }: StructureSelectorProps) {
  const { hasPermission } = useAuth();
  const {
    register,
    setValue,
    formState: { errors },
  } = useFormContext<CreateEventFormData>();

  const structureMode = useWatch<CreateEventFormData>({ name: "structure.mode" });
  const canCreateStructure = hasPermission("create", "structure");

  // Get structure errors based on current mode
  const structureErrors = errors.structure;
  const selectError = structureMode === "select" && structureErrors && "structureId" in structureErrors;
  const createNameError = structureMode === "create" && structureErrors && "name" in structureErrors;

  /**
   * Handle mode change with proper form value reset
   */
  const handleModeChange = (mode: "select" | "create") => {
    if (mode === "select") {
      setValue("structure", { mode: "select", structureId: 0 }, { shouldValidate: false });
    } else {
      setValue(
        "structure",
        {
          mode: "create",
          name: "",
          blinds: [{ small: 25, big: 50, ante: 0, time: 15 }],
        },
        { shouldValidate: false },
      );
    }
  };

  // Transform structures to Select options
  const structureOptions = structures.map((s) => ({
    value: String(s.id),
    label: s.name,
  }));

  return (
    <div className={styles.container}>
      <h3 className={styles.sectionTitle}>Structure</h3>

      <div className={styles.radioGroup}>
        <Radio
          name="structureMode"
          value="select"
          checked={structureMode === "select"}
          onChange={() => handleModeChange("select")}
          disabled={structures.length === 0 && !isLoadingStructures}
          label="Select existing structure"
          data-qa="radio-structure-mode-select"
        />

        {canCreateStructure && (
          <Radio
            name="structureMode"
            value="create"
            checked={structureMode === "create"}
            onChange={() => handleModeChange("create")}
            label="Create new structure"
            data-qa="radio-structure-mode-create"
          />
        )}
      </div>

      <div className={styles.content}>
        {structureMode === "select" ? (
          <>
            {isLoadingStructures ? (
              <div className={styles.loadingContainer} data-qa="structures-loading">
                <Spinner size="sm" />
                <span>Loading structures...</span>
              </div>
            ) : structures.length === 0 ? (
              <p className={styles.emptyMessage} data-qa="structures-empty">
                No structures available.{" "}
                {canCreateStructure ? "Please create a new structure." : "Contact an admin to create structures."}
              </p>
            ) : (
              <FormField
                label="Select Structure"
                htmlFor="structure.structureId"
                required
                error={
                  selectError
                    ? (structureErrors as { structureId?: { message?: string } }).structureId?.message
                    : undefined
                }
              >
                {(props) => (
                  <Select
                    {...props}
                    {...register("structure.structureId", { valueAsNumber: true })}
                    options={structureOptions}
                    placeholder="Select a structure"
                    error={!!selectError}
                    fullWidth
                    data-qa="select-structureId"
                  />
                )}
              </FormField>
            )}
          </>
        ) : (
          <>
            <FormField
              label="Structure Name"
              htmlFor="structure.name"
              required
              error={createNameError ? (structureErrors as { name?: { message?: string } }).name?.message : undefined}
            >
              {(props) => (
                <Input
                  {...props}
                  {...register("structure.name")}
                  type="text"
                  placeholder="e.g., Standard 15-min levels"
                  error={!!createNameError}
                  fullWidth
                  data-qa="input-structure-name"
                />
              )}
            </FormField>

            <div className={styles.blindLevelsSection}>
              <label className={styles.blindLevelsLabel}>Blind Levels</label>
              <BlindLevelsTable />
            </div>
          </>
        )}
      </div>
    </div>
  );
}
