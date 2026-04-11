export async function listSecurityPending(_params?: unknown) {
  return { items: [] };
}

export async function respondSecurity(_params?: unknown) {
  return {};
}

export async function getSecuritySummary(_params?: unknown) {
  return {
    summary: {
      security_status: "normal" as const,
      pending_authorizations: 0,
    },
  };
}
