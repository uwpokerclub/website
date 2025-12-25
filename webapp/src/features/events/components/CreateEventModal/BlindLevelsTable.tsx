import { useFieldArray, useFormContext, FieldErrors } from "react-hook-form";
import { Button, Input } from "@uwpokerclub/components";
import { FaPlus, FaTrash } from "react-icons/fa";
import type { CreateEventFormData, BlindLevel } from "../../schemas/eventSchema";
import styles from "./BlindLevelsTable.module.css";

// Type for blind level field errors
type BlindLevelErrors = FieldErrors<BlindLevel>;

/**
 * BlindLevelsTable - Dynamic table for managing blind levels
 *
 * Uses react-hook-form's useFieldArray for dynamic field management.
 * Automatically suggests values for new levels based on existing levels.
 */
export function BlindLevelsTable() {
  const {
    control,
    register,
    formState: { errors },
  } = useFormContext<CreateEventFormData>();

  const { fields, append, remove } = useFieldArray({
    control,
    name: "structure.blinds",
  });

  // Get blind level errors - safely access when in create mode
  const getBlindsErrors = (): BlindLevelErrors[] | undefined => {
    const structureErrors = errors.structure;
    if (structureErrors && "blinds" in structureErrors) {
      return structureErrors.blinds as BlindLevelErrors[] | undefined;
    }
    return undefined;
  };

  const blindsErrors = getBlindsErrors();

  /**
   * Add a new blind level with smart defaults
   * - If 1 level exists: double the previous values
   * - If 2+ levels exist: calculate progression based on last two levels
   */
  const handleAddLevel = () => {
    if (fields.length === 0) {
      append({ small: 25, big: 50, ante: 0, time: 15 });
      return;
    }

    const lastLevel = fields[fields.length - 1];

    if (fields.length === 1) {
      append({
        small: (lastLevel.small ?? 0) * 2,
        big: (lastLevel.big ?? 0) * 2,
        ante: (lastLevel.ante ?? 0) * 2,
        time: lastLevel.time ?? 15,
      });
      return;
    }

    const secondLastLevel = fields[fields.length - 2];
    const smallDiff = (lastLevel.small ?? 0) - (secondLastLevel.small ?? 0);
    const bigDiff = (lastLevel.big ?? 0) - (secondLastLevel.big ?? 0);
    const anteDiff = (lastLevel.ante ?? 0) - (secondLastLevel.ante ?? 0);

    const newSmall = Math.max(0, (lastLevel.small ?? 0) + smallDiff);
    const newBig = Math.max(newSmall, (lastLevel.big ?? 0) + bigDiff);
    const newAnte = Math.max(0, (lastLevel.ante ?? 0) + anteDiff);

    append({
      small: newSmall,
      big: newBig,
      ante: newAnte,
      time: lastLevel.time ?? 15,
    });
  };

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <span className={styles.headerCell}>Small</span>
        <span className={styles.headerCell}>Big</span>
        <span className={styles.headerCell}>Ante</span>
        <span className={styles.headerCell}>Time (min)</span>
      </div>

      {fields.map((field, index) => (
        <div key={field.id} className={styles.levelContainer}>
          <div className={styles.row}>
            <div className={styles.cell}>
              <Input
                {...register(`structure.blinds.${index}.small`, { valueAsNumber: true })}
                type="number"
                min={0}
                error={!!blindsErrors?.[index]?.small}
                fullWidth
              />
            </div>
            <div className={styles.cell}>
              <Input
                {...register(`structure.blinds.${index}.big`, { valueAsNumber: true })}
                type="number"
                min={0}
                error={!!blindsErrors?.[index]?.big}
                fullWidth
              />
            </div>
            <div className={styles.cell}>
              <Input
                {...register(`structure.blinds.${index}.ante`, { valueAsNumber: true })}
                type="number"
                min={0}
                error={!!blindsErrors?.[index]?.ante}
                fullWidth
              />
            </div>
            <div className={styles.cell}>
              <Input
                {...register(`structure.blinds.${index}.time`, { valueAsNumber: true })}
                type="number"
                min={1}
                error={!!blindsErrors?.[index]?.time}
                fullWidth
              />
            </div>
          </div>
          {fields.length > 1 && (
            <button type="button" className={styles.removeButton} onClick={() => remove(index)} title="Remove level">
              <FaTrash size={12} /> Remove
            </button>
          )}
        </div>
      ))}

      <Button type="button" variant="secondary" onClick={handleAddLevel} iconBefore={<FaPlus />}>
        Add Level
      </Button>
    </div>
  );
}
