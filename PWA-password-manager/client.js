class PasswordManagerApplication {
  constructor() {
    this.dbName = 'PasswordManagerDB';
    this.dbVersion = 1;
    this.tables = {
      profiles: 'profiles',
      domains: 'domains',
    };
    this.db = null;
  }
  OpenDB() {
    const request = window.indexedDB.open(this.dbName, this.dbVersion);
    const { profiles, domains } = this.tables;
    request.onupgradeneeded = (event) => {
      const db = event.target.result;
      if (!db.objectStoreNames.contains(domains)) {
        const domainTable = db.createObjectStore(domains, {
          keyPath: 'id',
          autoIncrement: true,
        });
        domainTable.createIndex('name', 'name', { unique: true });
      }
      if (!db.objectStoreNames.contains(profiles)) {
        const profileTable = db.createObjectStore(profiles, {
          keyPath: 'id',
          autoIncrement: true,
        });

        profileTable.createIndex('domainId', 'domainId', { unique: false });
      }
    };

    request.onsuccess = async (event) => {
      this.db = event.target.result;
      console.log('Database opened successfully');
      await this.runApplication();
    };

    request.onerror = (event) => {
      console.error('Database error:', event.target.error);
    };
  }
  async runApplication() {
    await this.displayList();
  }

  async getAllDomains() {
    return new Promise((res, rej) => {
      const transaction = this.db.transaction('domains', 'readwrite');
      const domainTable = transaction.objectStore('domains');
      const getAllRequeset = domainTable.getAll();
      getAllRequeset.onsuccess = () => {
        if (getAllRequeset.result) {
          res(getAllRequeset.result);
        } else {
          res([]);
        }
      };
      getAllRequeset.onerror = () => {
        rej(getAllRequeset.error);
      };
    });
  }

  async createDomain(name) {
    return new Promise((res, rej) => {
      const transaction = this.db.transaction('domains', 'readwrite');
      const domainTable = transaction.objectStore('domains');
      const addRequest = domainTable.add({
        name: name,
      });
      addRequest.onsuccess = () => {
        res({
          id: addRequest.result,
          name: name,
        });
      };
      addRequest.onerror = () => {
        res(false);
      };
    });
  }
  async deleteDomain(id) {
    return new Promise((res, rej) => {
      const transaction = this.db.transaction('domains', 'readwrite');
      const domainTable = transaction.objectStore('domains');
      const deleteRequest = domainTable.delete(Number(id));

      deleteRequest.onsuccess = async () => {
        const result = await this.deleteProfilesByDomainId(id);
        if (result) {
          res(true);
        }
      };
      deleteRequest.onerror = () => {
        res(false);
      };
    });
  }

  async getProfileById(id) {
    return new Promise((res, rej) => {
      const transaction = this.db.transaction('profiles', 'readonly');
      const profileTable = transaction.objectStore('profiles');
      const getRequest = profileTable.get(Number(id));
      getRequest.onsuccess = () => {
        res(getRequest.result);
      };
      getRequest.onerror = () => {
        rej(getRequest.error);
      };
    });
  }

  async getDoMainById(id) {
    return new Promise((res, rej) => {
      const transaction = this.db.transaction('domains', 'readonly');
      const domainTable = transaction.objectStore('domains');
      const getRequest = domainTable.get(Number(id));
      getRequest.onsuccess = () => {
        res(getRequest.result);
      };
      getRequest.onerror = () => {
        rej(getRequest.error);
      };
    });
  }

  async getProfilesByDomain(domainId) {
    return new Promise((res, rej) => {
      const transaction = this.db.transaction('profiles', 'readonly');
      const profileTable = transaction.objectStore('profiles');
      const index = profileTable.index('domainId');
      const getRequest = index.openCursor(IDBKeyRange.only(Number(domainId)));
      const profiles = [];
      getRequest.onsuccess = () => {
        const cursor = getRequest.result;
        if (cursor) {
          profiles.push(cursor.value);
          cursor.continue();
        } else {
          res(profiles);
        }
      };
      getRequest.onerror = () => {
        rej(getRequest.error);
      };
    });
  }

  async createProfileByDomainId(user) {
    const transaction = this.db.transaction('profiles', 'readwrite');
    const profileTable = transaction.objectStore('profiles');
    return new Promise((res, rej) => {
      const timestamp = Date.now().valueOf();
      const { username, password, domainId } = user;

      const profile = {
        username: username,
        password: password,
        domainId: domainId,
        updatedAt: timestamp,
        oldPassword: [],
      };
      const addRequest = profileTable.add(profile);
      addRequest.onsuccess = () => {
        profile.id = addRequest.result;
        res(profile);
      };
      addRequest.onerror = () => {
        rej(addRequest.error);
      };
    });
  }

  async deleteProfilesByDomainId(domainId) {
    return new Promise((res, rej) => {
      const transaction = this.db.transaction('profiles', 'readwrite');
      const profileTable = transaction.objectStore('profiles');
      const index = profileTable.index('domainId');
      const deleteRequest = index.openCursor(
        IDBKeyRange.only(Number(domainId))
      );
      let deletedCount = 0;
      deleteRequest.onsuccess = (event) => {
        const cursor = event.target.result;

        if (cursor) {
          const deleteOp = cursor.delete();
          deleteOp.onsuccess = () => {
            deletedCount++;
          };
          cursor.continue();
        } else {
          console.log(
            `Deleted ${deletedCount} profiles with domainId: ${domainId}`
          );
          res(true);
        }
      };

      deleteRequest.onerror = (event) => {
        console.error('Error deleting profiles:', event.target.error);
        rej(event.target.error);
      };
    });
  }

  async updatePassword(profile, newPassword) {
    return new Promise((res, rej) => {
      const transaction = this.db.transaction('profiles', 'readwrite');
      const profileTable = transaction.objectStore('profiles');
      const timestamp = Date.now().valueOf();
      const { password, updatedAt } = profile;
      profile.password = newPassword;
      profile.updatedAt = timestamp;
      profile.oldPassword.push({
        password: password,
        savedAt: updatedAt,
      });
      const putRequest = profileTable.put(profile);
      putRequest.onsuccess = () => {
        console.log('updated');
        res(profile);
      };
      putRequest.onerror = () => {
        rej(putRequest.error);
      };
    });
  }

  async deleteProfile(id) {
    return new Promise((res, rej) => {
      const transaction = this.db.transaction('profiles', 'readwrite');
      const profileTable = transaction.objectStore('profiles');
      const delRequest = profileTable.delete(Number(id));
      delRequest.onsuccess = () => {
        console.log('Deleted profile');
        res(true);
      };
      delRequest.onerror = () => {
        res(false);
      };
    });
  }

  generatePassword() {
    var length = 8,
      charset =
        'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789',
      retVal = '';
    for (var i = 0, n = charset.length; i < length; ++i) {
      retVal += charset.charAt(Math.floor(Math.random() * n));
    }
    return retVal;
  }
  async displayList() {
    //Get element (input,buttons,box)
    const domainsDiv = document.querySelector('#domains-list');
    const infoBox = document.querySelector('#info-box');
    const infoBoxContent = document.querySelector('#info-box-content');
    const deleteDomainBox = document.querySelector('#delete-domain-box');
    const createDomainBox = document.querySelector('#create-domain-box');
    const createProfileBox = document.querySelector('#create-profile-box');
    const domainNameInput = document.querySelector('#input-create-domain');
    const passwordCreateInput = document.querySelector(
      '#input-password-create-profile'
    );
    const loginCreateInput = document.querySelector(
      '#input-username-create-profile'
    );

    const confirmProfileCreateButton = createProfileBox.querySelector(
      '#profile-create-button'
    );
    //global variable to delay click (avoid double click)
    let debounceTimeout;

    //get all domains with profiles
    const allDomains = await this.getAllDomains();
    if (allDomains.length > 0) {
      let jsx = '';
      //and create a JSX string
      for (const domain of allDomains) {
        const { id, name } = domain;
        const profiles = await this.getProfilesByDomain(id);

        let profileJSX = '';
        if (profiles.length > 0) {
          profileJSX = profiles.reduce((accumulator, profile) => {
            const { username, password, id, updatedAt } = profile;
            const parsedDate = new Date(updatedAt).toLocaleString('en-US');
            return (
              accumulator +
              `<div class="grid grid-cols-4 pt-2 border-t-2 border-black border-solid ">
                <div class="flex flex-col  col-span-4 scale-[80%]">
                  <div>login: ${username}</div>
                  <div class="flex flex-row">password:${' '}
                  <input class="max-w-[170px] focus:outline-none" id="${
                    'input' + name + username + id
                  }" type="password" value=${password} disabled></div>
                  <div  class="flex gap-0 scale-[80%]">
                    <div class="">updated at:</div> 
                    <div class="" id=${
                      'date' + name + username + id
                    }>${parsedDate}
                    </div>
                  </div>
                </div>
                <div data-input=${
                  'input' + name + username + id
                } class="toggle-password w-full scale-75 bg-gray-500 text-white px-1 text-center ">show</div>
                <div data-user-id=${id} data-input=${
                'input' + name + username + id
              } data-date=${'date' + name + username + id} 
              class="edit-profile w-full scale-75 bg-gray-500 text-white px-1 text-center ">edit</div>
                <div data-user-id=${id} class="info-profile w-full scale-75 bg-sky-500 text-white px-1 text-center ">info</div>
                <div data-user-id=${id} class="delete-profile w-full scale-75 bg-red-500 text-white px-1 text-center ">delete</div>
              </div>`
            );
          }, '');
        }

        jsx += `
      <div id=${id} class="sm:w-[70%] w-[85%] flex flex-col m-auto border-black border-2 justify-center mb-2">
        <div class="bg-gray-400">
        <div class="">
          <p class="text-center scale-100 p-1">${name}</p>
        </div>
        <div class="grid grid-cols-2">
          <div domain-id=${id} class="add-profile bg-black w-full text-white text-center cursor-pointer scale-75 p-1">Add profile</div>
          <div domain-id=${id} class="delete-domain bg-yellow-600 w-full text-white text-center cursor-pointer scale-75 p-1">Delete domain</div>
        </div>
        </div>
        <div class="flex flex-col">${profileJSX}</div>
      </div>
    `;
      }
      //implement that JSX string into html file
      domainsDiv.innerHTML = jsx;
    } else {
      domainsDiv.innerHTML = '';
    }
    //Handle client behavior(click)
    const handleCloseBox = (box) => {
      box.classList.add('hidden');
      document.body.style.overflow = 'auto';
    };

    const domClickHandler = async (event) => {
      event.stopPropagation();
      if (debounceTimeout) return;
      debounceTimeout = setTimeout(() => {
        debounceTimeout = null;
      }, 10);
      const target = event.target;

      if (target.classList.contains('delete-domain')) {
        const domainId = target.getAttribute('domain-id');
        deleteDomainBox.classList.remove('hidden');
        document.body.style.overflow = 'hidden';
        document
          .getElementById('domain-delete-button')
          .setAttribute('domain-id', domainId);
      }
      if (target.classList.contains('toggle-password')) {
        const inputId = target.getAttribute('data-input');
        const passwordInput = document.getElementById(inputId);
        if (passwordInput) {
          passwordInput.type =
            passwordInput.type === 'password' ? 'text' : 'password';
          target.innerHTML = target.innerHTML === 'show' ? 'hide' : 'show';
        }
      }
      if (target.classList.contains('edit-profile')) {
        const profileId = target.getAttribute('data-user-id');
        const inputId = target.getAttribute('data-input');
        const dateId = target.getAttribute('data-date');
        const dateField = document.getElementById(dateId);
        const passwordInput = document.getElementById(inputId);
        if (passwordInput) {
          target.innerHTML = target.innerHTML === 'edit' ? 'save' : 'edit';
          passwordInput.disabled = passwordInput.disabled ? false : true;
          if (!passwordInput.disabled) {
            passwordInput.setAttribute('old-value', passwordInput.value);
            passwordInput.classList.add('text-green-600');
            target.classList.add('bg-green-600');
            passwordInput.focus();
          } else {
            passwordInput.classList.remove('text-green-600');
            target.classList.remove('bg-green-600');
            const oldValue = passwordInput.getAttribute('old-value');
            const newValue = passwordInput.value;
            if (oldValue === newValue) {
              console.log('no change in password, no update');
            } else {
              const profile = await this.getProfileById(profileId);
              const update = await this.updatePassword(profile, newValue);
              if (update) {
                const parsedDate = new Date(update.updatedAt).toLocaleString(
                  'en-US'
                );
                dateField.innerHTML = parsedDate;
              }
            }
          }
        }
      }
      if (target.classList.contains('delete-profile')) {
        console.log('deleting');
        const profileId = target.getAttribute('data-user-id');
        const deleteProfile = await this.deleteProfile(profileId);
        if (deleteProfile) {
          document.removeEventListener('click', domClickHandler);
          await this.displayList();
        }
      }
      if (target.classList.contains('info-profile')) {
        console.log('showing info');
        const profileId = target.getAttribute('data-user-id');
        const profile = await this.getProfileById(profileId);
        const domain = await this.getDoMainById(profile.domainId);
        const passwordList = profile.oldPassword.reduce((accumulator, item) => {
          const { password, savedAt } = item;
          const parsedDate = new Date(savedAt).toLocaleString('en-US');
          accumulator += `
          <div  class="flex">
            <div class="italic pr-2">${parsedDate}:</div>
            <div>${password}</div>
          </div>`;
          return accumulator;
        }, '');

        const content = `
        <div class="text-center">
          domain: ${domain.name}
        </div>
        <div> login: ${profile.username}</div>
        <div> password (current): ${profile.password}</div>
        <div class="max-h-[120px]">
        <div> password record:</div>
        ${passwordList || 'no history found for this profile'}
        </div>
        `;
        infoBoxContent.innerHTML = content;
        infoBox.classList.remove('hidden');
        document.body.style.overflow = 'hidden';
      }
      if (target.classList.contains('add-profile')) {
        const domainId = target.getAttribute('domain-id');
        createProfileBox.classList.remove('hidden');
        document.body.style.overflow = 'hidden';
        loginCreateInput.focus();
        confirmProfileCreateButton.setAttribute('domain-id', domainId);
      }
      if (target.id === 'profile-create-button') {
        const domainId = target.getAttribute('domain-id');
        if (loginCreateInput.value !== '' && passwordCreateInput.value !== '') {
          const newprofile = {
            username: loginCreateInput.value.replaceAll(' ', ''),
            password: passwordCreateInput.value,
            domainId: Number(domainId),
          };
          passwordCreateInput.value = '';
          loginCreateInput.value = '';
          const createProfile = await this.createProfileByDomainId(newprofile);
          if (createProfile) {
            document.removeEventListener('click', domClickHandler);
            await this.displayList();
            handleCloseBox(createProfileBox);
          } else {
            handleCloseBox(createProfileBox);
          }
        } else {
          alert('Please fill all fields');
          // handleCloseBox(createProfileBox);
        }
      }
      if (target.id === 'profile-not-create-button') {
        handleCloseBox(createProfileBox);
      }
      if (target.id === 'info-close-button') {
        handleCloseBox(infoBox);
      }
      if (target.id === 'domain-not-create-button') {
        handleCloseBox(createDomainBox);
      }
      if (target.id === 'add-new-domain') {
        createDomainBox.classList.remove('hidden');
        document.body.style.overflow = 'hidden';
        domainNameInput.focus();
      }
      if (target.id === 'domain-not-delete-button') {
        deleteDomainBox.classList.add('hidden');
        document.body.style.overflow = 'hidden';
      }
      if (target.id === 'domain-delete-button') {
        const domainId = target.getAttribute('domain-id');
        const deleteDomain = await this.deleteDomain(domainId);
        if (deleteDomain) {
          document.removeEventListener('click', domClickHandler);
          await this.displayList();
        }
        deleteDomainBox.classList.add('hidden');
      }
      if (target.id === 'domain-create-button') {
        if (domainNameInput.value !== '') {
          const newDomain = await this.createDomain(
            domainNameInput.value.replaceAll(' ', '')
          );
          domainNameInput.value = '';
          if (newDomain) {
            document.removeEventListener('click', domClickHandler);
            await this.displayList();
            handleCloseBox(createDomainBox);
          } else {
            console.log('this domain has already existed');
            alert('this domain has already existed');
            handleCloseBox(createDomainBox);
          }
        }
      }
      if (target.id === 'password-generate-button') {
        const generatedPassword = this.generatePassword();
        console.log(generatedPassword);
        passwordCreateInput.value = generatedPassword;
      }
    };
    document.addEventListener('click', domClickHandler);
  }
}

if ('serviceWorker' in navigator) {
  navigator.serviceWorker
    .register('./service-worker.js')
    .then(() => console.log('Service Worker registered'))
    .catch((err) => console.error('Service Worker registration failed:', err));
}

let deferredPrompt;

window.addEventListener('beforeinstallprompt', (e) => {
  e.preventDefault();
  deferredPrompt = e;
  console.log('Install prompt event fired');

  const installButton = document.getElementById('install-button');
  installButton.classList.remove('hidden');
  const handleInstallClick = () => {
    deferredPrompt.prompt();
    deferredPrompt.userChoice.then((choiceResult) => {
      if (choiceResult.outcome === 'accepted') {
        installButton.classList.add('hidden');
        console.log('User accepted the install prompt');
      } else {
        console.log('User dismissed the install prompt');
      }
      deferredPrompt = null;
      installButton.removeEventListener('click', handleInstallClick);
    });
  };
  if (installButton) {
    installButton.addEventListener('click', handleInstallClick);
  }
});

const passwordManager = new PasswordManagerApplication();
passwordManager.OpenDB();
