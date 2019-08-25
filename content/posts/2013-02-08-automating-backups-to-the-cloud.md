---
'@context': http://schema.org
'@type': BlogPosting
datePublished: "2013-02-08"
description: Combining gpg, Amazon S3 and IAM policies.
keywords:
- backup
- gpg
- s3
name: Automating Backups to the Cloud
url:
- /blog/2013/02/08/automating-backups-to-the-cloud.html
---

Backups are extremely important and I've been experimenting with a few different methods. My concerns are always focused
on maintaining data integrity, security, and availability. One of my current methods involves using asymmetric keys for
secure storage and object versioning to ensure backup data can't undesirably be overwritten.


## Encryption Keys

For encryption and decryption I'm using asymmetric keys via [`gpg`][1]. This way, any server can generate and encrypt
the data, but only administrators who have the private key could actually decrypt the data. Generating the
administrative key looks like:

```
$ gpg --gen-key
gpg (GnuPG) 1.4.11; Copyright (C) 2010 Free Software Foundation, Inc.
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.

... [snip] ...

gpg: key CEFAF45B marked as ultimately trusted
public and secret key created and signed.

gpg: checking the trustdb
gpg: 3 marginal(s) needed, 1 complete(s) needed, PGP trust model
gpg: depth: 0  valid:   4  signed:   0  trust: 0-, 0q, 0n, 0m, 0f, 4u
pub   2048R/CEFAF45B 2013-02-08
      Key fingerprint = 46DF 2951 7E2D 41D7 F7B5  EB16 20C2 1C03 CEFA F45B
uid                  Danny Berger (secret-project-backup) <dpb587@gmail.com>
sub   2048R/765C4556 2013-02-08
```

To actually use the public key on servers, it can be exported and copied...

```
$ gpg --armor --export 'Danny Berger (secret-project-backup) <dpb587@gmail.com>'
-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v1.4.11 (Darwin)
... [snip] ...
-----END PGP PUBLIC KEY BLOCK-----
```

Then pasted and imported on the machine(s) that will be encrypting data...

```
$ cat | gpg --import
gpg: directory `/home/app-devtools/.gnupg' created
gpg: new configuration file `/home/app-devtools/.gnupg/gpg.conf' created
gpg: WARNING: options in `/home/app-devtools/.gnupg/gpg.conf' are not yet active during this run
gpg: keyring `/home/app-devtools/.gnupg/secring.gpg' created
gpg: keyring `/home/app-devtools/.gnupg/pubring.gpg' created
... [snip] ... Ctrl-D ...
gpg: /home/app-devtools/.gnupg/trustdb.gpg: trustdb created
gpg: key CEFAF45B: public key "Danny Berger (secret-project-backup) <dpb587@gmail.com>" imported
gpg: Total number processed: 1
gpg:               imported: 1  (RSA: 1)
```

And then marked as "ultimately trusted" with the `trust` command (otherwise it always wants to confirm before using the
key)...

```
$ gpg --edit-key 'Danny Berger (secret-project-backup) <dpb587@gmail.com>'
... [snip] ...

pub  2048R/CEFAF45B  created: 2013-02-08  expires: never       usage: SC  
                     trust: ultimate      validity: unknown
sub  2048R/765C4556  created: 2013-02-08  expires: never       usage: E   
[ unknown] (1). Danny Berger (secret-project-backup) <dpb587@gmail.com>
Please note that the shown key validity is not necessarily correct
unless you restart the program.

Command> quit
```


## Amazon S3

In my case, I wanted to regularly send the encrypted backups offsite and [S3][2] seemed like a flexible, effective
storage place. This involved a couple steps:

**Create a new S3 bucket** (e.g. `backup.secret-project.example.com`) - this will just hold all the different backup
types and files for the project.

**Enable Object Versioning** on the S3 bucket - whenever a new backup gets dropped off, previous backups will remain.
This provides for additional security (e.g. a compromised server could not overwrite the backup with an empty file) and
more complex retention policies than Amazon's Glacier lifecycle rules.

**Create a new IAM user** (e.g. `com-example-secret-project-backup`) - the user and it's Access Key will be responsible
for uploading the backup files to the bucket.

**Add a User Policy** to the IAM user - the only permission it needs is `PutObject` for the bucket:

```javascript
{
  "Statement" : [
    {
      "Sid" : "Stmt0123456789",
      "Action" : [
        "s3:PutObject"
      ],
      "Effect" : "Allow",
      "Resource" : [
        "arn:aws:s3:::backup.secret-project.example.com/*"
      ]
    }
  ]
}
```


**Upload Method** - instead of depending on third-party libraries for uploading the backup files, I wanted to try simply
using `curl` with Amazon S3's Browser-Based Upload functionality. This involved creating and signing the appropriate
policy via the [sample][3] policy builder for a particular backup type. My simple policy looked like:

```javascript
{
  "expiration" : "2016-01-01T12:00:00.000Z",
  "conditions" : [
    { "bucket" : "backup.secret-project.example.com" },
    { "acl" : "private" },
    { "key" : "database.sql.gz.enc" },
  ]
}
```


## All Together

Putting everything together, a single command could be used to backup the database, compress, encrypt, and upload:

```
$ mysqldump ... \
    | gzip -c \
    | gpg --recipient 'Danny Berger (secret-project-backup) <dpb587@gmail.com>' --encrypt \
    | curl \
        -F key=database.sql.gz.enc \
        -F acl=private \
        -F AWSAccessKeyId=AKIA99076E3F28E55AF85 \
        -F policy=ewogICJleHBpcmF0aW9uIiA6ICIyMDE2LTAxLTAxVDEyOjAwOjAwLjAwMFoiLAogICJjb25kaXRpb25zIiA6IFsKICAgIHsgImJ1Y2tldCIgOiAiYmFja3VwLnNlY3JldC1wcm9qZWN0LmV4YW1wbGUuY29tIiB9LAogICAgeyAiYWNsIiA6ICJwcml2YXRlIiB9LAogICAgeyAia2V5IiA6ICJkYXRhYmFzZS5zcWwuZ3ouZW5jIiB9LAogIF0KfQ== \
        -F signature=937ca778e4d44db7b804cfdd70d= \
        -F file=@- \
        https://s3.amazonaws.com/backup.secret-project.example.com
```

And then to download, decrypt, decompress, and reload the database from an administrative machine:

```
$ wget -qO- 'https://s3.amazonaws.com/backup.secret-project.example.com/database.sql.tgz.enc?versionId=c0b55912f42c4142bcb44c3eb1376f35&AWSAccessKeyId=AKIA99076E3F28E55AF85&...' \
    | gpg -d \
    | gunzip \
    | mysql ...
```

The only task remaining is creating a cleanup script using the S3 API to monitor the different backup versions and
delete them as they expire.


## Summary

While it has a bit of overhead to get things set up properly, using `gpg` makes secure backups trivial and S3 provides
the flexible storage strategy to ensure data is safe.


 [1]: http://www.gnupg.org/
 [2]: http://aws.amazon.com/s3/
 [3]: http://s3.amazonaws.com/doc/s3-example-code/post/post_sample.html
