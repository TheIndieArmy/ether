import { Badge, Button, SettingsSection, toast } from "@braedonsaunders/appkit-ui";
import { useState } from "react";
import { Field, TestButton, type TestState } from "../../components/fields";
import { SettingsRow } from "../../components/SettingsRow";
import { api, ApiError } from "../../lib/api";
import type { ArrConfig, ArrInstance } from "../../lib/types";
import { t } from "../../i18n";

type Props = {
  value: ArrConfig;
  saved?: ArrConfig;
  onChange: (next: ArrConfig) => void;
  errors: Record<string, string>;
  footer?: React.ReactNode;
};

type Probe = {
  rootFolders?: { id: number; path: string }[];
  qualityProfiles?: { id: number; name: string }[];
  message?: string;
};

/**
 * Integrations › Arr: master enable unlocks Sonarr/Radarr connection forms.
 * A successful Test loads root folders and quality profiles for defaults.
 */
export function ArrSettings({ value, onChange, errors }: Props) {
  const set = (patch: Partial<ArrConfig>) => onChange({ ...value, ...patch });
  const setInst = (kind: "radarr" | "sonarr", patch: Partial<ArrInstance>) =>
    set({ [kind]: { ...value[kind], ...patch } });

  return (
    <div className="space-y-6">
      <SettingsSection
        title={t("Arr integration")}
        description={t(
          "Connect Sonarr and Radarr so Ether can prefer downloaded files on their root folders, and optionally request missing titles.",
        )}
      >
        <SettingsRow
          title={t("Enable Arrs")}
          description={t("Turn this on to unlock Sonarr and Radarr connection settings.")}
        >
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={Boolean(value.enabled)}
              onChange={(e) => set({ enabled: e.target.checked })}
              aria-label={t("Enable Arrs")}
            />
            {value.enabled ? <Badge variant="success">{t("On")}</Badge> : <Badge variant="secondary">{t("Off")}</Badge>}
          </label>
        </SettingsRow>
      </SettingsSection>

      {value.enabled ? (
        <>
          <ArrInstanceCard
            kind="radarr"
            label="Radarr"
            value={value.radarr}
            errorPrefix="arr.radarr"
            errors={errors}
            onChange={(inst) => setInst("radarr", inst)}
          />
          <ArrInstanceCard
            kind="sonarr"
            label="Sonarr"
            value={value.sonarr}
            errorPrefix="arr.sonarr"
            errors={errors}
            onChange={(inst) => setInst("sonarr", inst)}
          />
          <SettingsSection
            title={t("Path maps")}
            description={t(
              "Rewrite Arr container paths to paths this Ether host can read (Docker binds).",
            )}
          >
            {(value.pathMaps ?? []).map((m, i) => (
              <div key={i} className="grid gap-2 sm:grid-cols-2">
                <Field label={t("Arr path prefix")} error={errors[`arr.pathMaps[${i}].arrPrefix`]}>
                  <input
                    className="w-full rounded border border-border bg-surface px-2 py-1.5 text-sm"
                    value={m.arrPrefix}
                    onChange={(e) => {
                      const pathMaps = [...(value.pathMaps ?? [])];
                      pathMaps[i] = {
                        arrPrefix: e.target.value,
                        localPrefix: pathMaps[i]?.localPrefix ?? "",
                      };
                      set({ pathMaps });
                    }}
                  />
                </Field>
                <Field label={t("Local path prefix")} error={errors[`arr.pathMaps[${i}].localPrefix`]}>
                  <input
                    className="w-full rounded border border-border bg-surface px-2 py-1.5 text-sm"
                    value={m.localPrefix}
                    onChange={(e) => {
                      const pathMaps = [...(value.pathMaps ?? [])];
                      pathMaps[i] = {
                        arrPrefix: pathMaps[i]?.arrPrefix ?? "",
                        localPrefix: e.target.value,
                      };
                      set({ pathMaps });
                    }}
                  />
                </Field>
              </div>
            ))}
            <Button
              type="button"
              variant="secondary"
              size="sm"
              onClick={() => set({ pathMaps: [...(value.pathMaps ?? []), { arrPrefix: "", localPrefix: "" }] })}
            >
              {t("Add path map")}
            </Button>
          </SettingsSection>
        </>
      ) : null}
    </div>
  );
}

function ArrInstanceCard({
  kind,
  label,
  value,
  errorPrefix,
  errors,
  onChange,
}: {
  kind: "radarr" | "sonarr";
  label: string;
  value: ArrInstance;
  errorPrefix: string;
  errors: Record<string, string>;
  onChange: (patch: Partial<ArrInstance>) => void;
}) {
  const [test, setTest] = useState<TestState>({ status: "idle" });
  const [probe, setProbe] = useState<Probe | null>(null);

  const runTest = async () => {
    setTest({ status: "running" });
    try {
      const r = await api.testArr(kind, value);
      if (!r.ok) {
        setTest({ status: "fail", message: r.error ?? t("failed") });
        setProbe(null);
        return;
      }
      setTest({ status: "ok", durationMs: r.durationMs });
      setProbe((r.probe as Probe) ?? null);
      toast.success(t("{app} connected", { app: label }));
    } catch (e) {
      setTest({ status: "fail", message: e instanceof ApiError ? e.message : (e as Error).message });
      setProbe(null);
    }
  };

  return (
    <SettingsSection title={label}>
      <Field label={t("URL")} error={errors[`${errorPrefix}.url`]}>
        <input
          className="w-full rounded border border-border bg-surface px-2 py-1.5 text-sm"
          value={value.url}
          placeholder={kind === "radarr" ? "http://radarr:7878" : "http://sonarr:8989"}
          onChange={(e) => onChange({ url: e.target.value })}
          aria-label={`${label} URL`}
        />
      </Field>
      <Field label={t("API key")} error={errors[`${errorPrefix}.apiKey`]}>
        <input
          type="password"
          className="w-full rounded border border-border bg-surface px-2 py-1.5 text-sm"
          value={value.apiKey}
          autoComplete="off"
          onChange={(e) => onChange({ apiKey: e.target.value })}
          aria-label={`${label} API key`}
        />
      </Field>
      <TestButton label={t("Test {app}", { app: label })} state={test} onTest={() => void runTest()} />
      {probe ? (
        <div className="space-y-3 rounded-lg border border-border p-3">
          <p className="text-xs text-fg-muted">{probe.message}</p>
          <Field label={t("Default root folder")}>
            <select
              className="w-full rounded border border-border bg-surface px-2 py-1.5 text-sm"
              value={value.rootFolderPath}
              onChange={(e) => onChange({ rootFolderPath: e.target.value })}
            >
              <option value="">{t("Select…")}</option>
              {(probe.rootFolders ?? []).map((r) => (
                <option key={r.id} value={r.path}>
                  {r.path}
                </option>
              ))}
            </select>
          </Field>
          <Field label={t("Default quality profile")}>
            <select
              className="w-full rounded border border-border bg-surface px-2 py-1.5 text-sm"
              value={value.qualityProfileId || ""}
              onChange={(e) =>
                onChange({ qualityProfileId: e.target.value === "" ? 0 : Number(e.target.value) })
              }
            >
              <option value="">{t("Select…")}</option>
              {(probe.qualityProfiles ?? []).map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                </option>
              ))}
            </select>
          </Field>
        </div>
      ) : null}
    </SettingsSection>
  );
}
