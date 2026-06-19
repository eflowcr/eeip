import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [CommonModule],
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
  // Using localhost:10000 for local testing, in prod it should be an env variable or relative path
  private apiUrl = 'http://localhost:10000/api/v1';

  ngOnInit() {
    this.loadImportantEmails();
    // The following arrays are initialized empty. 
    // They will be hydrated once the respective backend endpoints are implemented.
    this.inbox = [];
    this.risks = [];
    this.commitments = [];
    this.contacts = [];
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
}
