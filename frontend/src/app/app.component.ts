import { Component, OnInit, inject, ChangeDetectorRef, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { FormsModule } from '@angular/forms';
import { AuthService } from './auth.service';
import { Login } from './login/login';
import { environment } from '../environments/environment';
import { ChartOptions } from './chart-types';
import { NgApexchartsModule } from 'ng-apexcharts';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [CommonModule, FormsModule, Login, NgApexchartsModule],
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements OnInit {
  authService = inject(AuthService);
  
  get isAdmin() {
    return this.authService.currentUser()?.role === 'Admin';
  }

  get isAuditor() {
    return this.authService.currentUser()?.role === 'Auditor';
  }

  get canAudit() {
    return this.isAdmin || this.isAuditor;
  }

  constructor() {
    effect(() => {
      if (this.authService.currentUser()) {
        this.loadImportantEmails();
        if (this.isAdmin) {
          this.loadAccounts();
          this.loadStakeholders();
          this.loadUsers();
        }
      }
    });
  }

  logout() {
    this.authService.logout();
    this.activeTab = 'dashboard';
  }

  title = 'EEIP - Plataforma de Inteligencia Ejecutiva';
  importantEmails: any[] = [];
  pendingEmails: any[] = [];
  auditingEmails: any[] = [];
  closedEmails: any[] = [];

  // Filters
  searchQuery: string = '';
  selectedAccountFilter: string = '';

  get uniqueMonitoredAccounts() {
    const accounts = this.importantEmails.map(e => e.monitored_account);
    return [...new Set(accounts)].filter(Boolean);
  }

  filterEmails(emails: any[]) {
    return emails.filter(e => {
      const q = this.searchQuery.toLowerCase();
      const matchSearch = q === '' || 
        (e.sender_email && e.sender_email.toLowerCase().includes(q)) || 
        (e.subject && e.subject.toLowerCase().includes(q));
      const matchAccount = this.selectedAccountFilter === '' || 
        e.monitored_account === this.selectedAccountFilter;
      return matchSearch && matchAccount;
    });
  }

  get filteredPendingEmails() { return this.filterForDashboard(this.filterEmails(this.pendingEmails)); }
  get filteredAuditingEmails() { return this.filterForDashboard(this.filterEmails(this.auditingEmails)); }
  get filteredClosedEmails() { return this.filterForDashboard(this.filterEmails(this.closedEmails)); }
  
  // Tones for filtering
  get activeTones(): string[] {
    const tones = new Set<string>();
    this.importantEmails.forEach(e => {
      if (e.detected_tone) tones.add(e.detected_tone);
    });
    return Array.from(tones);
  }

  // Priorities for filtering
  get activePriorities(): string[] {
    const priorities = new Set<string>();
    this.importantEmails.forEach(e => {
      if (e.priority) priorities.add(e.priority);
    });
    return Array.from(priorities);
  }

  selectedToneFilter: string | null = null;
  selectedPriorityFilter: string | null = null;

  toggleToneFilter(tone: string) {
    this.selectedToneFilter = this.selectedToneFilter === tone ? null : tone;
  }

  togglePriorityFilter(priority: string) {
    this.selectedPriorityFilter = this.selectedPriorityFilter === priority ? null : priority;
  }

  filterForDashboard(emails: any[]) {
    return emails.filter(e => {
      let matchTone = true;
      let matchPriority = true;
      
      if (this.selectedToneFilter) {
        matchTone = e.detected_tone?.toLowerCase() === this.selectedToneFilter.toLowerCase();
      }
      if (this.selectedPriorityFilter) {
        matchPriority = e.priority?.toLowerCase() === this.selectedPriorityFilter.toLowerCase();
      }
      return matchTone && matchPriority;
    });
  }
  
  inbox: any[] = [];
  risks: any[] = [];
  commitments: any[] = [];
  contacts: any[] = [];
  activeTab = 'dashboard';

  // Charts
  public isQuantumChartCollapsed = false;
  public isFiltersCollapsed = false;
  public dashboardChartOptions!: Partial<ChartOptions>;
  public stakeholdersChartOptions!: Partial<ChartOptions>;

  private http = inject(HttpClient);
  private cdr = inject(ChangeDetectorRef);
  private apiUrl = environment.apiUrl;

  // Configuración de cuenta
  accounts: any[] = [];
  newAccount: any = {
    account_name: '',
    email_address: '',
    imap_host: '',
    imap_port: 993,
    imap_user: '',
    imap_password: ''
  };
  isSavingAccount = false;
  isTestingConnection = false;
  accountSuccessMessage = '';
  accountErrorMessage = '';
  editingAccountId: string | null = null;
  
  // Para pruebas en línea en la lista
  accountTestStatus: { [id: string]: { loading: boolean, message: string, error: boolean } } = {};
  accountSyncStatus: { [id: string]: { loading: boolean, message: string, error: boolean } } = {};
  // Stakeholders
  stakeholders: any[] = [];
  newStakeholder: any = {
    name: '',
    email: '',
    telegram_chat_id: ''
  };
  isSavingStakeholder = false;

  // Usuarios
  users: any[] = [];

  ngOnInit() {
    this.inbox = [];
    this.risks = [];
    this.commitments = [];
    this.contacts = [];
  }

  loadAccounts() {
    this.http.get<any[]>(`${this.apiUrl}/accounts`).subscribe({
      next: (data) => {
        if (data) {
          this.accounts = data;
          this.cdr.detectChanges();
        }
      },
      error: (err) => {
        console.error('Error cargando cuentas', err);
      }
    });
  }

  loadStakeholders() {
    this.http.get<any[]>(`${this.apiUrl}/stakeholders`).subscribe({
      next: (data) => {
        if (data) {
          this.stakeholders = data;
          this.cdr.detectChanges();
        }
      },
      error: (err) => {
        console.error('Error cargando interesados', err);
      }
    });
  }

  loadUsers() {
    this.http.get<any[]>(`${this.apiUrl}/users`).subscribe({
      next: (data) => {
        if (data) {
          this.users = data;
          this.cdr.detectChanges();
        }
      },
      error: (err) => {
        console.error('Error cargando usuarios', err);
      }
    });
  }

  updateUserRole(user: any, newRole: string) {
    this.http.put(`${this.apiUrl}/users/${user.id}/role`, { role: newRole }).subscribe({
      next: () => {
        user.role = newRole;
        this.cdr.detectChanges();
      },
      error: (err) => {
        console.error('Error actualizando rol de usuario', err);
        // Refresh users to revert the select box if it failed
        this.loadUsers();
      }
    });
  }

  deleteUser(userId: string) {
    if (confirm('¿Estás seguro de eliminar a este usuario del sistema?')) {
      this.http.delete(`${this.apiUrl}/users/${userId}`).subscribe({
        next: () => {
          this.users = this.users.filter(u => u.id !== userId);
          this.cdr.detectChanges();
        },
        error: (err) => {
          console.error('Error eliminando usuario', err);
        }
      });
    }
  }

  saveStakeholder() {
    if (!this.newStakeholder.name || !this.newStakeholder.email) return;
    this.isSavingStakeholder = true;
    this.http.post(`${this.apiUrl}/stakeholders`, this.newStakeholder).subscribe({
      next: () => {
        this.isSavingStakeholder = false;
        this.newStakeholder = { name: '', email: '', telegram_chat_id: '' };
        this.loadStakeholders();
      },
      error: (err) => {
        this.isSavingStakeholder = false;
        console.error('Error guardando interesado', err);
      }
    });
  }

  deleteStakeholder(id: string) {
    if (!confirm('¿Estás seguro de eliminar a este interesado?')) return;
    this.http.delete(`${this.apiUrl}/stakeholders/${id}`).subscribe({
      next: () => {
        this.loadStakeholders();
      },
      error: (err) => {
        console.error('Error eliminando interesado', err);
      }
    });
  }

  loadImportantEmails() {
    this.http.get<any[]>(`${this.apiUrl}/emails/important?limit=50`).subscribe({
      next: (data) => {
        if (data) {
          const today = new Date();
          const isToday = (dateString: string) => {
            const date = new Date(dateString);
            return date.getDate() === today.getDate() &&
                   date.getMonth() === today.getMonth() &&
                   date.getFullYear() === today.getFullYear();
          };

          this.importantEmails = data;
          this.pendingEmails = data.filter(e => (e.status === 'Unread' || e.status === 'Read' || !e.status) && (this.canAudit || !e.user_seen));
          this.auditingEmails = data.filter(e => e.status === 'Auditing');
          this.closedEmails = data.filter(e => e.status === 'Actioned' && isToday(e.updated_at));
          this.risks = data.filter(e => (e.customer_risk_score && e.customer_risk_score > 50) || (e.escalation_risk_score && e.escalation_risk_score > 50) || e.priority === 'Critical');
          this.commitments = data.filter(e => e.requires_action === true);

          const contactsMap = new Map<string, any>();
          this.importantEmails.forEach(e => {
            if (!e.sender_email) return;
            const sender = e.sender_email;
            if (!contactsMap.has(sender)) {
              contactsMap.set(sender, {
                email: sender,
                totalEmails: 0,
                toneCounts: {} as { [key: string]: number },
                predominantTone: 'Neutral',
                lastContactAt: e.received_at
              });
            }
            const contact = contactsMap.get(sender);
            contact.totalEmails++;
            const tone = e.detected_tone || 'Neutral';
            contact.toneCounts[tone] = (contact.toneCounts[tone] || 0) + 1;
            if (new Date(e.received_at) > new Date(contact.lastContactAt)) {
              contact.lastContactAt = e.received_at;
            }
          });

          this.contacts = Array.from(contactsMap.values()).map(contact => {
            let maxCount = 0;
            let predTone = 'Neutral';
            for (const [tone, count] of Object.entries(contact.toneCounts)) {
              if ((count as number) > maxCount) {
                maxCount = count as number;
                predTone = tone;
              }
            }
            contact.predominantTone = predTone;
            // Also store array of tones for the UI to loop over
            contact.topTones = Object.entries(contact.toneCounts)
              .map(([t, c]) => ({ tone: t, count: c }))
              .sort((a, b) => (b.count as number) - (a.count as number));
            return contact;
          }).sort((a, b) => b.totalEmails - a.totalEmails);

          this.initCharts();

          this.cdr.detectChanges();
        }
      },
      error: (err) => {
        console.error('Error cargando correos importantes', err);
      }
    });
  }

  getCriticalCount(): number {
    return this.pendingEmails.filter(e => e.priority === 'Critical').length + this.auditingEmails.filter(e => e.priority === 'Critical').length;
  }

  isLoadingInbox = false;

  loadInbox() {
    this.isLoadingInbox = true;
    this.cdr.detectChanges();
    this.http.get<any[]>(`${this.apiUrl}/emails/all?limit=100`).subscribe({
      next: (data) => {
        this.isLoadingInbox = false;
        if (data) {
          this.inbox = data;
        } else {
          this.inbox = [];
        }
        this.cdr.detectChanges();
      },
      error: (err) => {
        this.isLoadingInbox = false;
        console.error('Error cargando bandeja global', err);
        this.cdr.detectChanges();
      }
    });
  }

  inboxSearchSender: string = '';
  inboxSearchRecipient: string = '';
  inboxFilterCategory: string = '';
  inboxFilterPriority: string = '';
  inboxFilterTone: string = '';

  get activeInboxCategories(): string[] {
    const cats = new Set<string>();
    this.inbox.forEach(e => {
      if (e.category) cats.add(e.category);
    });
    return Array.from(cats).sort();
  }

  get activeInboxPriorities(): string[] {
    const prios = new Set<string>();
    this.inbox.forEach(e => {
      if (e.priority) prios.add(e.priority);
    });
    return Array.from(prios).sort();
  }

  get activeInboxTones(): string[] {
    const tones = new Set<string>();
    this.inbox.forEach(e => {
      if (e.detected_tone) tones.add(e.detected_tone);
    });
    return Array.from(tones).sort();
  }

  get filteredInbox() {
    return this.inbox.filter(e => {
      let matchSender = true;
      if (this.inboxSearchSender) {
        matchSender = e.sender_email?.toLowerCase().includes(this.inboxSearchSender.toLowerCase());
      }
      
      let matchRecipient = true;
      if (this.inboxSearchRecipient) {
        matchRecipient = e.monitored_account?.toLowerCase().includes(this.inboxSearchRecipient.toLowerCase());
      }

      let matchCategory = true;
      if (this.inboxFilterCategory) {
        matchCategory = e.category === this.inboxFilterCategory;
      }

      let matchPriority = true;
      if (this.inboxFilterPriority) {
        matchPriority = e.priority === this.inboxFilterPriority;
      }

      let matchTone = true;
      if (this.inboxFilterTone) {
        matchTone = e.detected_tone === this.inboxFilterTone;
      }

      return matchSender && matchRecipient && matchCategory && matchPriority && matchTone;
    });
  }

  setTab(tab: string) {
    this.activeTab = tab;
    if (tab === 'dashboard') {
      this.loadImportantEmails();
    } else if (tab === 'inbox') {
      this.loadInbox();
    }
  }

  loadingSummaryId: string | null = null;

  toggleExpand(email: any) {
    email.expanded = !email.expanded;
  }

  isCritical(email: any): boolean {
    const isCriticalPriority = email.priority === 'Critical' || email.priority === 'Urgent';
    const t = email.detected_tone;
    const isNegativeTone = t === 'Agresivo/violento' || t === 'Amenazante' || t === 'Confrontativo' || t === 'Frustrado' || t === 'Angry' || t === 'Molesto';
    return isCriticalPriority || isNegativeTone;
  }

  generateSummary(email: any, event: Event) {
    event.stopPropagation();
    if (email.summary) {
      return; // Already generated
    }
    this.loadingSummaryId = email.id;
    this.cdr.detectChanges();

    this.http.post<{summary: string}>(`${this.apiUrl}/emails/${email.id}/summary`, {}).subscribe({
      next: (res) => {
        email.summary = res.summary;
        this.loadingSummaryId = null;
        this.importantEmails = [...this.importantEmails];
        this.cdr.detectChanges();
      },
      error: (err) => {
        console.error('Error generating summary', err);
        email.summary = 'No se pudo generar el resumen en este momento.';
        this.loadingSummaryId = null;
        this.importantEmails = [...this.importantEmails];
        this.cdr.detectChanges();
      }
    });
  }

  markAsAuditing(emailId: string, event: Event) {
    event.stopPropagation();
    this.http.put(`${this.apiUrl}/emails/${emailId}/status`, { status: 'Auditing' }).subscribe({
      next: () => {
        const email = this.importantEmails.find(e => e.id === emailId);
        if (email) email.status = 'Auditing';
        this.pendingEmails = this.pendingEmails.filter(e => e.id !== emailId);
        if (email) this.auditingEmails = [email, ...this.auditingEmails];
        this.cdr.detectChanges();
      },
      error: (err) => console.error('Error marking as auditing', err)
    });
  }

  markAsSeen(emailId: string, event: Event) {
    event.stopPropagation();
    this.http.put(`${this.apiUrl}/emails/${emailId}/seen`, {}).subscribe({
      next: () => {
        const email = this.importantEmails.find(e => e.id === emailId);
        if (email) email.user_seen = true;
        if (!this.canAudit) {
          this.pendingEmails = this.pendingEmails.filter(e => e.id !== emailId);
        }
        this.cdr.detectChanges();
      },
      error: (err) => console.error('Error marking as seen', err)
    });
  }

  markAsResolved(emailId: string, event: Event) {
    event.stopPropagation();
    this.http.put(`${this.apiUrl}/emails/${emailId}/status`, { status: 'Actioned' }).subscribe({
      next: () => {
        const email = this.importantEmails.find(e => e.id === emailId);
        if (email) email.status = 'Actioned';
        this.pendingEmails = this.pendingEmails.filter(e => e.id !== emailId);
        this.auditingEmails = this.auditingEmails.filter(e => e.id !== emailId);
        if (email) this.closedEmails = [email, ...this.closedEmails];
        this.cdr.detectChanges();
      },
      error: (err) => console.error('Error marking as resolved', err)
    });
  }

  saveAccount() {
    this.isSavingAccount = true;
    this.accountSuccessMessage = '';
    this.accountErrorMessage = '';

    const req = this.editingAccountId 
      ? this.http.put(`${this.apiUrl}/accounts/${this.editingAccountId}`, this.newAccount)
      : this.http.post(`${this.apiUrl}/accounts`, this.newAccount);

    req.subscribe({
      next: (res) => {
        this.isSavingAccount = false;
        this.accountSuccessMessage = this.editingAccountId ? '¡Cuenta actualizada exitosamente!' : '¡Cuenta configurada y guardada exitosamente!';
        this.resetAccountForm();
        this.loadAccounts();
        this.cdr.detectChanges();
      },
      error: (err) => {
        this.isSavingAccount = false;
        this.accountErrorMessage = 'Error al guardar la cuenta. Revisa la conexión con el servidor.';
        this.cdr.detectChanges();
      }
    });
  }

  resetAccountForm() {
    this.newAccount = { account_name: '', email_address: '', imap_host: '', imap_port: 993, imap_user: '', imap_password: '', is_private: false };
    this.editingAccountId = null;
  }

  editAccount(acc: any) {
    this.editingAccountId = acc.id;
    this.newAccount = { ...acc, imap_password: '' }; // Don't prefill password
    this.accountSuccessMessage = '';
    this.accountErrorMessage = '';
  }

  deleteAccount(id: string) {
    if (confirm('¿Estás seguro de que deseas eliminar esta cuenta y detener su monitoreo?')) {
      this.http.delete(`${this.apiUrl}/accounts/${id}`).subscribe({
        next: () => {
          this.loadAccounts();
        },
        error: (err) => {
          alert('Error al eliminar la cuenta');
        }
      });
    }
  }

  testConnection(acc: any = null) {
    const dataToTest = acc || this.newAccount;
    if (!dataToTest.imap_host || !dataToTest.imap_user || !dataToTest.imap_password) {
      this.accountErrorMessage = 'Completa los campos del servidor, usuario y contraseña para probar la conexión.';
      return;
    }

    this.isTestingConnection = true;
    this.accountSuccessMessage = '';
    this.accountErrorMessage = '';

    this.http.post(`${this.apiUrl}/accounts/test`, dataToTest).subscribe({
      next: () => {
        this.isTestingConnection = false;
        this.accountSuccessMessage = '¡Conexión IMAP exitosa! Las credenciales son válidas.';
        this.cdr.detectChanges();
      },
      error: (err) => {
        this.isTestingConnection = false;
        const msg = err.error?.details || 'Credenciales inválidas o servidor inalcanzable.';
        this.accountErrorMessage = `Error IMAP: ${msg}`;
        this.cdr.detectChanges();
      }
    });
  }

  testExistingAccount(accountId: string) {
    this.accountTestStatus[accountId] = { loading: true, message: 'Probando...', error: false };
    this.cdr.detectChanges();

    this.http.post(`${this.apiUrl}/accounts/${accountId}/test`, {}).subscribe({
      next: () => {
        this.accountTestStatus[accountId] = { loading: false, message: '¡Conexión exitosa!', error: false };
        this.cdr.detectChanges();
        // Clear message after 3 seconds
        setTimeout(() => {
          if (this.accountTestStatus[accountId]) {
            this.accountTestStatus[accountId].message = '';
            this.cdr.detectChanges();
          }
        }, 3000);
      },
      error: (err) => {
        const msg = err.error?.details || 'Error de conexión';
        this.accountTestStatus[accountId] = { loading: false, message: `Error: ${msg}`, error: true };
        this.cdr.detectChanges();
      }
    });
  }

  syncAccount(accountId: string) {
    this.accountSyncStatus[accountId] = { loading: true, message: 'Sincronizando...', error: false };
    this.cdr.detectChanges();

    this.http.post(`${this.apiUrl}/accounts/${accountId}/sync`, {}).subscribe({
      next: () => {
        this.accountSyncStatus[accountId] = { loading: false, message: '¡Sincronización completada!', error: false };
        this.loadImportantEmails();
        this.cdr.detectChanges();
        setTimeout(() => {
          if (this.accountSyncStatus[accountId]) {
            this.accountSyncStatus[accountId].message = '';
            this.cdr.detectChanges();
          }
        }, 3000);
      },
      error: (err) => {
        const msg = err.error?.details || 'Error de sincronización';
        this.accountSyncStatus[accountId] = { loading: false, message: `Error: ${msg}`, error: true };
        this.cdr.detectChanges();
      }
    });
  }

  initCharts() {
    const priorities = ['Crítico', 'Alto', 'Medio', 'Bajo', 'Critical', 'High', 'Medium', 'Low'];
    const pGroups = ['Crítico', 'Alto', 'Medio', 'Bajo'];
    const tonesMap = new Map<string, number[]>();

    const getPriorityGroupIdx = (p: string) => {
      if (!p) return 3;
      const lp = p.toLowerCase();
      if (lp.includes('crític') || lp.includes('critic')) return 0;
      if (lp.includes('alto') || lp.includes('high')) return 1;
      if (lp.includes('medio') || lp.includes('medium')) return 2;
      return 3;
    };

    this.importantEmails.forEach(e => {
      const pIdx = getPriorityGroupIdx(e.priority);
      const tone = e.detected_tone || 'Neutral';
      if (!tonesMap.has(tone)) {
        tonesMap.set(tone, [0, 0, 0, 0]);
      }
      tonesMap.get(tone)![pIdx]++;
    });

    const dashboardSeries = Array.from(tonesMap.entries()).map(([tone, data]) => ({
      name: tone,
      data: data
    }));

    this.dashboardChartOptions = {
      series: dashboardSeries,
      chart: { type: 'bar', height: 350, stacked: true, background: 'transparent', toolbar: { show: false }, dropShadow: { enabled: true, color: '#000', top: 0, left: 0, blur: 5, opacity: 0.3 } },
      colors: ['#6366f1', '#10b981', '#d946ef', '#06b6d4', '#f59e0b', '#f43f5e'],
      plotOptions: { bar: { horizontal: false, borderRadius: 4, columnWidth: '40%' } },
      dataLabels: { enabled: false },
      stroke: { width: 0, colors: ['transparent'] },
      xaxis: { categories: pGroups, labels: { style: { colors: '#9ca3af' } } },
      yaxis: { labels: { style: { colors: '#9ca3af' } } },
      legend: { position: 'top', labels: { colors: '#9ca3af' } },
      grid: { borderColor: '#374151', strokeDashArray: 4 },
      tooltip: { theme: 'dark' }
    };

    const stkTonesMap = new Map<string, number[]>();
    const stakeholderNames = this.contacts.slice(0, 10).map(c => c.email);

    this.contacts.slice(0, 10).forEach((c, idx) => {
      c.topTones.forEach((t: any) => {
        const tone = t.tone;
        if (!stkTonesMap.has(tone)) {
          stkTonesMap.set(tone, Array(10).fill(0));
        }
        stkTonesMap.get(tone)![idx] = t.count;
      });
    });

    const stakeholdersSeries = Array.from(stkTonesMap.entries()).map(([tone, data]) => ({
      name: tone,
      data: data.slice(0, stakeholderNames.length)
    }));

    this.stakeholdersChartOptions = {
      series: stakeholdersSeries,
      chart: { type: 'bar', height: 350, stacked: true, background: 'transparent', toolbar: { show: false }, dropShadow: { enabled: true, color: '#000', top: 0, left: 0, blur: 5, opacity: 0.3 } },
      colors: ['#06b6d4', '#d946ef', '#10b981', '#6366f1', '#f59e0b', '#f43f5e'],
      plotOptions: { bar: { horizontal: true, borderRadius: 4, barHeight: '50%' } },
      dataLabels: { enabled: false },
      stroke: { width: 0, colors: ['transparent'] },
      xaxis: { categories: stakeholderNames, labels: { style: { colors: '#9ca3af' } } },
      yaxis: { labels: { style: { colors: '#9ca3af' } } },
      legend: { position: 'top', labels: { colors: '#9ca3af' } },
      grid: { borderColor: '#374151', strokeDashArray: 4 },
      tooltip: { theme: 'dark' }
    };
  }
}
