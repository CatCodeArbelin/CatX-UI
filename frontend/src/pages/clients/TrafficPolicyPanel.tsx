import { useCallback, useEffect, useState } from 'react';
import { Button, Card, Descriptions, InputNumber, Space, Switch, Tag, message } from 'antd';
import { useTranslation } from 'react-i18next';

import { HttpUtil, SizeFormatter } from '@/utils';

interface TrafficPolicyView {
  enabled: boolean;
  windowSeconds: number;
  quotaBytes: number;
  activeUploadBps: number;
  activeDownloadBps: number;
  throttleUploadBps: number;
  throttleDownloadBps: number;
  lifecycle: string;
  reason: string;
  usedBytes: number;
  remainingBytes: number;
  enforcement: string;
  enforcementNote?: string;
}

interface TrafficPolicyPanelProps {
  email: string;
}

const EMPTY_POLICY: TrafficPolicyView = {
  enabled: false,
  windowSeconds: 3600,
  quotaBytes: 0,
  activeUploadBps: 0,
  activeDownloadBps: 0,
  throttleUploadBps: 0,
  throttleDownloadBps: 0,
  lifecycle: 'active',
  reason: '',
  usedBytes: 0,
  remainingBytes: 0,
  enforcement: 'unsupported',
  enforcementNote: 'No policy configured',
};

export default function TrafficPolicyPanel({ email }: TrafficPolicyPanelProps) {
  const { t } = useTranslation();
  const [view, setView] = useState<TrafficPolicyView | null>(null);
  const [loading, setLoading] = useState(false);
  const [resetting, setResetting] = useState(false);
  const [saving, setSaving] = useState(false);
  const [draft, setDraft] = useState<Partial<TrafficPolicyView>>({});
  const [messageApi, contextHolder] = message.useMessage();

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const response = (await HttpUtil.get(
        `/panel/api/traffic-control/clients/${encodeURIComponent(email)}/policy`,
        undefined,
        { silent: true },
      )) as { success?: boolean; obj?: TrafficPolicyView };
      const next = response.success ? response.obj ?? null : null;
      setView(next ?? EMPTY_POLICY);
      if (next) setDraft(next);
    } finally {
      setLoading(false);
    }
  }, [email]);

  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);

  async function reset() {
    setResetting(true);
    try {
      const response = (await HttpUtil.post(
        `/panel/api/traffic-control/clients/${encodeURIComponent(email)}/policy/reset`,
      )) as { success?: boolean; obj?: TrafficPolicyView; msg?: string };
      if (!response.success) {
        messageApi.error(response.msg || t('somethingWentWrong'));
        return;
      }
      setView(response.obj ?? null);
    } finally {
      setResetting(false);
    }
  }

  async function save() {
    setSaving(true);
    try {
      const response = (await HttpUtil.put(
        `/panel/api/traffic-control/clients/${encodeURIComponent(email)}/policy`,
        draft,
      )) as { success?: boolean; obj?: TrafficPolicyView; msg?: string };
      if (!response.success) {
        messageApi.error(response.msg || t('somethingWentWrong'));
        return;
      }
      setView(response.obj ?? null);
      if (response.obj) setDraft(response.obj);
    } finally {
      setSaving(false);
    }
  }

  if (!view) return null;
  const tagColor = view.enforcement === 'unsupported' ? 'orange' : view.enforcement === 'degraded' ? 'red' : 'green';
  return (
    <>
      {contextHolder}
      <Card
        size="small"
        title={t('pages.inbounds.traffic')}
        loading={loading}
        style={{ marginTop: 16 }}
        extra={<Button size="small" type="primary" onClick={() => void save()} loading={saving}>{t('save')}</Button>}
      >
        <Descriptions size="small" column={1}>
          <Descriptions.Item label={t('enabled')}>
            <Switch checked={Boolean(draft.enabled)} onChange={(enabled) => setDraft((v) => ({ ...v, enabled }))} />
          </Descriptions.Item>
          <Descriptions.Item label={t('pages.inbounds.traffic')}>
            <Space wrap>
              <InputNumber min={60} value={draft.windowSeconds} onChange={(windowSeconds) => setDraft((v) => ({ ...v, windowSeconds: windowSeconds ?? 3600 }))} addonAfter="s" />
              <InputNumber min={0} value={draft.quotaBytes} onChange={(quotaBytes) => setDraft((v) => ({ ...v, quotaBytes: quotaBytes ?? 0 }))} addonAfter="bytes" />
            </Space>
          </Descriptions.Item>
          <Descriptions.Item label={t('pages.clients.speed')}>
            <Space wrap>
              <InputNumber min={0} value={draft.activeUploadBps} onChange={(activeUploadBps) => setDraft((v) => ({ ...v, activeUploadBps: activeUploadBps ?? 0 }))} addonAfter="↑ B/s" />
              <InputNumber min={0} value={draft.activeDownloadBps} onChange={(activeDownloadBps) => setDraft((v) => ({ ...v, activeDownloadBps: activeDownloadBps ?? 0 }))} addonAfter="↓ B/s" />
            </Space>
          </Descriptions.Item>
          <Descriptions.Item label={t('pages.clients.speed')}>
            <Space wrap>
              <InputNumber min={0} value={draft.throttleUploadBps} onChange={(throttleUploadBps) => setDraft((v) => ({ ...v, throttleUploadBps: throttleUploadBps ?? 0 }))} addonAfter="↑ B/s" />
              <InputNumber min={0} value={draft.throttleDownloadBps} onChange={(throttleDownloadBps) => setDraft((v) => ({ ...v, throttleDownloadBps: throttleDownloadBps ?? 0 }))} addonAfter="↓ B/s" />
            </Space>
          </Descriptions.Item>
          <Descriptions.Item label={t('status')}>
            <Tag color={view.lifecycle === 'throttled' ? 'orange' : view.lifecycle === 'disabled' ? 'red' : 'green'}>
              {view.lifecycle}
            </Tag>
            {view.reason ? <span className="hint">{view.reason}</span> : null}
          </Descriptions.Item>
          <Descriptions.Item label={t('status')}>
            <Tag color={tagColor}>{view.enforcement}</Tag>
            {view.enforcementNote ? <span className="hint">{view.enforcementNote}</span> : null}
          </Descriptions.Item>
          <Descriptions.Item label={t('pages.inbounds.traffic')}>
            {SizeFormatter.sizeFormat(view.usedBytes)} / {view.quotaBytes > 0 ? SizeFormatter.sizeFormat(view.quotaBytes) : '∞'}
          </Descriptions.Item>
          <Descriptions.Item label={t('remaining')}>
            {view.quotaBytes > 0 ? SizeFormatter.sizeFormat(view.remainingBytes) : '∞'}
          </Descriptions.Item>
        </Descriptions>
        <Button size="small" onClick={() => void reset()} loading={resetting} disabled={view.lifecycle !== 'throttled'}>{t('reset')}</Button>
      </Card>
    </>
  );
}
