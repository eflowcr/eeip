import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements OnInit {
  title = 'EEIP - Plataforma de Inteligencia Ejecutiva';
  importantEmails: any[] = [];
  inbox: any[] = [];
  risks: any[] = [];
  commitments: any[] = [];
  contacts: any[] = [];
  activeTab = 'dashboard';

  private http = inject(HttpClient);
  private apiUrl = 'http://localhost:10000/api/v1';

  // Configuración de cuenta
  accounts: any[] = [];
  newAccount = {
    email_address: '',
    imap_host: '',
    imap_port: 993,
    imap_user: '',
    imap_password: ''
  };
  isSavingAccount = false;
  accountSuccessMessage = '';

  ngOnInit() {
    this.loadImportantEmails();
    this.loadAccounts();
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
        }
      },
      error: (err) => {
        console.error('Error cargando cuentas', err);
      }
    });
  }

  loadImportantEmails() {
    this.http.get<any[]>(`${this.apiUrl}/emails/important?limit=10`).subscribe({
      next: (data) => {
        if (data) {
          this.importantEmails = data;
        }
      },
      error: (err) => {
        console.error('Error cargando correos importantes', err);
      }
    });
  }

  setTab(tab: string) {
    this.activeTab = tab;
  }

  saveAccount() {
    this.isSavingAccount = true;
    this.accountSuccessMessage = '';
    this.http.post(`${this.apiUrl}/accounts`, this.newAccount).subscribe({
      next: (res) => {
        this.isSavingAccount = false;
        this.accountSuccessMessage = '¡Cuenta configurada y guardada exitosamente!';
        this.newAccount = { email_address: '', imap_host: '', imap_port: 993, imap_user: '', imap_password: '' };
        this.loadAccounts();
      },
      error: (err) => {
        this.isSavingAccount = false;
        alert('Error al guardar la cuenta. Revisa la conexión con el servidor.');
      }
    });
  }
}
